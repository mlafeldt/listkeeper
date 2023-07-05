package main

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/mlafeldt/listkeeper/functions/internal/data"
	"github.com/mlafeldt/listkeeper/functions/internal/evb"
	"github.com/mlafeldt/listkeeper/functions/internal/twitter"
)

var ignoreFollowerEventFields = cmpopts.IgnoreFields(data.FollowerEvent{}, "ID", "CreatedAt", "ExpiresAt")

type tableStub struct {
	data.TableAPI

	lists           []*data.FollowerList
	ignoreFollowers []string
}

func (t *tableStub) GetUser(ctx context.Context, userID string) (*data.User, error) {
	user := data.User{
		ID:              userID,
		IgnoreFollowers: t.ignoreFollowers,
	}
	return &user, nil
}

func (t *tableStub) GetUserAndLatestFollowerLists(ctx context.Context, userID string, limit int64) (*data.User, []*data.FollowerList, error) {
	user, _ := t.GetUser(ctx, userID)
	return user, t.lists, nil
}

func (t *tableStub) CreateFollowerEvent(ctx context.Context, e *data.FollowerEvent) error {
	return nil
}

type s3DownloaderStub struct {
	s3manageriface.DownloaderAPI

	followerIDs map[string][]int64
}

func (d *s3DownloaderStub) DownloadWithContext(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput, f ...func(*s3manager.Downloader)) (int64, error) {
	buf, err := json.Marshal(d.followerIDs[*input.Key])
	if err != nil {
		return 0, err
	}
	n, err := w.WriteAt(buf, 0)
	return int64(n), err
}

type evbStub struct {
	evb.API
}

func (e *evbStub) Send(ctx context.Context, eventType string, events ...interface{}) error {
	return nil
}

type twitterStub struct {
	twitter.API

	users  map[int64]*twitter.User
	errors map[int64]error
}

func (t *twitterStub) UserByID(ctx context.Context, accessToken, accessSecret string, userID int64) (*twitter.User, error) {
	if e, ok := t.errors[userID]; ok {
		return nil, e
	}
	if u, ok := t.users[userID]; ok {
		return u, nil
	}
	return nil, data.ErrUserNotFound
}

func TestNoChanges(t *testing.T) {
	h := handler{
		table: &tableStub{
			lists: []*data.FollowerList{
				{S3Key: "/some/path"},
				{S3Key: "/some/path"},
			},
		},
	}

	want := &output{}

	got, err := h.handle(context.Background(), input{UserID: "000"})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func TestNewFollower(t *testing.T) {
	h := handler{
		table: &tableStub{
			lists: []*data.FollowerList{
				{S3Key: "/new/path", TotalFollowers: 2},
				{S3Key: "/old/path"},
			},
		},
		s3Downloader: &s3DownloaderStub{
			followerIDs: map[string][]int64{
				"/new/path": {111, 222},
				"/old/path": {111},
			},
		},
		evb: &evbStub{},
		twitter: &twitterStub{
			users: map[int64]*twitter.User{
				222: {Handle: "bob"},
			},
		},
	}

	want := &output{
		Events: []*data.FollowerEvent{
			{
				UserID:              "000",
				TotalFollowers:      2,
				Follower:            &twitter.User{Handle: "bob"},
				FollowerState:       data.FollowerStateNew,
				FollowerStateReason: data.FollowerStateReasonFollowed,
			},
		},
	}

	got, err := h.handle(context.Background(), input{UserID: "000"})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got, ignoreFollowerEventFields); diff != "" {
		t.Error(diff)
	}
}

func TestLostFollower(t *testing.T) {
	h := handler{
		table: &tableStub{
			lists: []*data.FollowerList{
				{S3Key: "/new/path", TotalFollowers: 1},
				{S3Key: "/old/path"},
			},
		},
		s3Downloader: &s3DownloaderStub{
			followerIDs: map[string][]int64{
				"/new/path": {111},
				"/old/path": {111, 222, 333, 444},
			},
		},
		evb: &evbStub{},
		twitter: &twitterStub{
			users: map[int64]*twitter.User{
				222: {Handle: "bob"},
			},
			errors: map[int64]error{
				333: twitter.ErrUserNotFound,
				444: twitter.ErrUserSuspended,
			},
		},
	}

	want := &output{
		Events: []*data.FollowerEvent{
			{
				UserID:              "000",
				TotalFollowers:      1,
				Follower:            &twitter.User{Handle: "bob"},
				FollowerState:       data.FollowerStateLost,
				FollowerStateReason: data.FollowerStateReasonUnfollowed,
			},
			{
				UserID:              "000",
				TotalFollowers:      1,
				Follower:            &twitter.User{ID: "333"},
				FollowerState:       data.FollowerStateLost,
				FollowerStateReason: data.FollowerStateReasonDeleted,
			},
			{
				UserID:              "000",
				TotalFollowers:      1,
				Follower:            &twitter.User{ID: "444"},
				FollowerState:       data.FollowerStateLost,
				FollowerStateReason: data.FollowerStateReasonSuspended,
			},
		},
	}

	got, err := h.handle(context.Background(), input{UserID: "000"})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got, ignoreFollowerEventFields); diff != "" {
		t.Error(diff)
	}
}

func TestNewAndLostFollower(t *testing.T) {
	h := handler{
		table: &tableStub{
			lists: []*data.FollowerList{
				{S3Key: "/new/path", TotalFollowers: 2},
				{S3Key: "/old/path"},
			},
		},
		s3Downloader: &s3DownloaderStub{
			followerIDs: map[string][]int64{
				"/new/path": {111, 222},
				"/old/path": {111, 333},
			},
		},
		evb: &evbStub{},
		twitter: &twitterStub{
			users: map[int64]*twitter.User{
				222: {Handle: "bob"},
				333: {Handle: "carlos"},
			},
		},
	}

	want := &output{
		Events: []*data.FollowerEvent{
			{
				UserID:              "000",
				TotalFollowers:      2,
				Follower:            &twitter.User{Handle: "bob"},
				FollowerState:       data.FollowerStateNew,
				FollowerStateReason: data.FollowerStateReasonFollowed,
			},
			{
				UserID:              "000",
				TotalFollowers:      2,
				Follower:            &twitter.User{Handle: "carlos"},
				FollowerState:       data.FollowerStateLost,
				FollowerStateReason: data.FollowerStateReasonUnfollowed,
			},
		},
	}

	got, err := h.handle(context.Background(), input{UserID: "000"})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got, ignoreFollowerEventFields); diff != "" {
		t.Error(diff)
	}
}

func TestIgnoreFollowers(t *testing.T) {
	h := handler{
		table: &tableStub{
			lists: []*data.FollowerList{
				{S3Key: "/new/path", TotalFollowers: 2},
				{S3Key: "/old/path"},
			},
			ignoreFollowers: []string{"111", "carlos", "@dan"},
		},
		s3Downloader: &s3DownloaderStub{
			followerIDs: map[string][]int64{
				"/new/path": {111, 222},
				"/old/path": {333, 444},
			},
		},
		evb: &evbStub{},
		twitter: &twitterStub{
			users: map[int64]*twitter.User{
				111: {ID: "111"},
				222: {ID: "222"},
				333: {Handle: "carlos"},
				444: {Handle: "dan"},
			},
		},
	}

	want := &output{
		Events: []*data.FollowerEvent{
			{
				UserID:              "000",
				TotalFollowers:      2,
				Follower:            &twitter.User{ID: "222"},
				FollowerState:       data.FollowerStateNew,
				FollowerStateReason: data.FollowerStateReasonFollowed,
			},
		},
	}

	got, err := h.handle(context.Background(), input{UserID: "000"})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got, ignoreFollowerEventFields); diff != "" {
		t.Error(diff)
	}
}
