package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/kelseyhightower/envconfig"

	"github.com/mlafeldt/listkeeper/functions/internal/data"
	"github.com/mlafeldt/listkeeper/functions/internal/twitter"
)

type input struct {
	UserID string
}

type handler struct {
	table      data.TableAPI
	tableTTL   time.Duration
	s3Uploader s3manageriface.UploaderAPI
	bucketName string
	twitter    twitter.API
}

func main() {
	var env struct {
		TableName      string        `envconfig:"TABLE_NAME" required:"true"`
		TableTTL       time.Duration `envconfig:"TABLE_TTL" required:"true"`
		BucketName     string        `envconfig:"BUCKET_NAME" required:"true"`
		ConsumerKey    string        `envconfig:"TWITTER_CONSUMER_KEY" required:"true"`
		ConsumerSecret string        `envconfig:"TWITTER_CONSUMER_SECRET" required:"true"`
	}
	envconfig.MustProcess("", &env)

	sess := session.Must(session.NewSession())
	h := handler{
		table:      data.NewTable(sess, env.TableName),
		tableTTL:   env.TableTTL,
		s3Uploader: s3manager.NewUploader(sess),
		bucketName: env.BucketName,
		twitter:    twitter.NewClient(env.ConsumerKey, env.ConsumerSecret),
	}

	lambda.Start(h.handle)
}

func (h *handler) handle(ctx context.Context, in input) (*data.FollowerList, error) {
	log.SetPrefix(in.UserID + " ")
	log.Printf("input = %+v", in)

	if in.UserID == "" {
		return nil, errors.New("user ID must be passed as input")
	}

	user, err := h.table.GetUser(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	followerIDs, err := h.twitter.FollowerIDs(ctx, user.AccessToken, user.AccessSecret)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(followerIDs); err != nil {
		return nil, err
	}

	var (
		digest = sha256.Sum256(buf.Bytes())
		s3Key  = fmt.Sprintf("user/%s/followers/%s", user.ID, hex.EncodeToString(digest[:]))
		now    = time.Now()
	)

	_, err = h.s3Uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(h.bucketName),
		Key:         aws.String(s3Key),
		ContentType: aws.String("application/json"),
		Body:        &buf,
	})
	if err != nil {
		return nil, err
	}

	list := data.FollowerList{
		UserID:         user.ID,
		S3Bucket:       h.bucketName,
		S3Key:          s3Key,
		TotalFollowers: len(followerIDs),
		CreatedAt:      now,
		ExpiresAt:      now.Add(h.tableTTL),
	}

	if err := h.table.CreateFollowerList(ctx, &list); err != nil {
		return nil, err
	}

	log.Printf("output = %+v", list)

	return &list, nil
}
