package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/kelseyhightower/envconfig"
	"github.com/segmentio/ksuid"

	"github.com/mlafeldt/listkeeper/functions/internal/data"
	"github.com/mlafeldt/listkeeper/functions/internal/evb"
	"github.com/mlafeldt/listkeeper/functions/internal/twitter"
)

const numListsToCompare = 2

type input struct {
	UserID string
}

type output struct {
	Events []*data.FollowerEvent `json:",omitempty"` //nolint:tagliatelle
}

type handler struct {
	table        data.TableAPI
	eventTTL     time.Duration
	evb          evb.API
	s3Downloader s3manageriface.DownloaderAPI
	twitter      twitter.API
}

func main() {
	var env struct {
		TableName       string        `envconfig:"TABLE_NAME" required:"true"`
		EventTTL        time.Duration `envconfig:"EVENT_TTL" default:"2160h"` // 90 days
		EventBusName    string        `envconfig:"EVENT_BUS_NAME" required:"true"`
		EventSourceName string        `envconfig:"EVENT_SOURCE_NAME" required:"true"`
		ConsumerKey     string        `envconfig:"TWITTER_CONSUMER_KEY" required:"true"`
		ConsumerSecret  string        `envconfig:"TWITTER_CONSUMER_SECRET" required:"true"`
	}
	envconfig.MustProcess("", &env)

	sess := session.Must(session.NewSession())
	h := handler{
		table:    data.NewConsistentTable(sess, env.TableName),
		eventTTL: env.EventTTL,
		evb: evb.NewClient(sess, &evb.Config{
			EventBusName:    env.EventBusName,
			EventSourceName: env.EventSourceName,
		}),
		s3Downloader: s3manager.NewDownloader(sess),
		twitter:      twitter.NewClient(env.ConsumerKey, env.ConsumerSecret),
	}

	lambda.Start(h.handle)
}

//nolint:cyclop,gocognit
func (h *handler) handle(ctx context.Context, in input) (*output, error) {
	if in.UserID == "" {
		return nil, errors.New("user ID must be passed as input")
	}

	log.SetPrefix(in.UserID + " ")
	log.Printf("input = %+v", in)

	user, followerLists, err := h.table.GetUserAndLatestFollowerLists(ctx, in.UserID, numListsToCompare)
	if err != nil {
		return nil, err
	}
	if len(followerLists) < numListsToCompare {
		log.Print("too few follower lists, skipping diff")
		return &output{}, nil
	}

	// Only download and compare follower lists if the key (content hash) has changed
	if followerLists[0].S3Key == followerLists[1].S3Key {
		log.Print("follower lists did not change")
		return &output{}, nil
	}

	var followerIDs [numListsToCompare][]int64

	for i := 0; i < numListsToCompare; i++ {
		var buf aws.WriteAtBuffer

		_, err := h.s3Downloader.DownloadWithContext(ctx, &buf, &s3.GetObjectInput{
			Bucket: aws.String(followerLists[i].S3Bucket),
			Key:    aws.String(followerLists[i].S3Key),
		})
		if err != nil {
			return nil, err
		}

		var ids []int64
		if err := json.Unmarshal(buf.Bytes(), &ids); err != nil {
			return nil, err
		}

		followerIDs[i] = ids
	}

	var (
		_, lostFollowers, newFollowers = diffInt64Slices(followerIDs[1], followerIDs[0])
		totalFollowers                 = followerLists[0].TotalFollowers
		seq                            = ksuid.Sequence{Seed: ksuid.New()}
	)

	events := make([]*data.FollowerEvent, 0, len(newFollowers)+len(lostFollowers))

	for _, id := range newFollowers {
		follower, err := h.twitter.UserByID(ctx, user.AccessToken, user.AccessSecret, id)
		if err != nil {
			if errors.Is(err, twitter.ErrUserNotFound) || errors.Is(err, twitter.ErrUserSuspended) {
				// Ignore new follower gone in the meantime
				continue
			}
			return nil, err
		}

		if user.IgnoresFollower(follower.ID, follower.Handle) {
			log.Printf("ignoring new follower: %+v", follower)
			continue
		}

		eid, _ := seq.Next()
		events = append(events, &data.FollowerEvent{
			ID:                  eid.String(),
			UserID:              user.ID,
			TotalFollowers:      totalFollowers,
			Follower:            follower,
			FollowerState:       data.FollowerStateNew,
			FollowerStateReason: data.FollowerStateReasonFollowed,
			CreatedAt:           eid.Time(),
			ExpiresAt:           eid.Time().Add(h.eventTTL),
		})
	}

	for _, id := range lostFollowers {
		reason := data.FollowerStateReasonUnfollowed

		follower, err := h.twitter.UserByID(ctx, user.AccessToken, user.AccessSecret, id)
		if err != nil {
			follower = &twitter.User{ID: strconv.FormatInt(id, 10)} //nolint:gomnd

			switch {
			case errors.Is(err, twitter.ErrUserNotFound):
				reason = data.FollowerStateReasonDeleted
			case errors.Is(err, twitter.ErrUserSuspended):
				reason = data.FollowerStateReasonSuspended
			default:
				return nil, err
			}
		}

		if user.IgnoresFollower(follower.ID, follower.Handle) {
			log.Printf("ignoring lost follower: %+v", follower)
			continue
		}

		eid, _ := seq.Next()
		events = append(events, &data.FollowerEvent{
			ID:                  eid.String(),
			UserID:              user.ID,
			TotalFollowers:      totalFollowers,
			Follower:            follower,
			FollowerState:       data.FollowerStateLost,
			FollowerStateReason: reason,
			CreatedAt:           eid.Time(),
			ExpiresAt:           eid.Time().Add(h.eventTTL),
		})
	}

	out := output{
		Events: events,
	}

	log.Printf("output = %+v", out)

	for _, e := range events {
		if err := h.table.CreateFollowerEvent(ctx, e); err != nil {
			return nil, err
		}
		if err := h.evb.Send(ctx, "Twitter Follower Change", e); err != nil {
			return nil, err
		}
	}

	return &out, nil
}
