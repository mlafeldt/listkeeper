package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/mlafeldt/listkeeper/functions/internal/data"
	"github.com/mlafeldt/listkeeper/functions/internal/twitter"
)

type tableStub struct {
	data.TableAPI

	user *data.User
}

func (t tableStub) GetUser(ctx context.Context, userID string) (*data.User, error) {
	return t.user, nil
}

func (t *tableStub) CreateFollowerList(ctx context.Context, l *data.FollowerList) error {
	return nil
}

type s3UploaderStub struct {
	s3manageriface.UploaderAPI
}

func (*s3UploaderStub) UploadWithContext(ctx context.Context, input *s3manager.UploadInput, f ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	return nil, nil //nolint:nilnil
}

type twitterStub struct {
	twitter.API

	followerIDs []int64
}

func (t *twitterStub) FollowerIDs(ctx context.Context, accessToken, accessSecret string) ([]int64, error) {
	return t.followerIDs, nil
}

func TestGetFollowers(t *testing.T) {
	h := handler{
		table: &tableStub{
			user: data.NewUser("000"),
		},
		s3Uploader: &s3UploaderStub{},
		bucketName: "some-bucket",
		twitter: &twitterStub{
			followerIDs: []int64{123, 456, 789},
		},
	}

	// $ echo '[123,456,789]' | sha256sum
	// 9b4620ebc5b5ebf2c51da1b778f141c22d7c48b68c74e7c3c3931d4a17894c81  -
	want := &data.FollowerList{
		UserID:         "000",
		S3Bucket:       "some-bucket",
		S3Key:          "user/000/followers/9b4620ebc5b5ebf2c51da1b778f141c22d7c48b68c74e7c3c3931d4a17894c81",
		TotalFollowers: 3,
	}

	got, err := h.handle(context.Background(), input{UserID: "000"})
	if err != nil {
		t.Fatal(err)
	}

	opts := cmpopts.IgnoreFields(data.FollowerList{}, "CreatedAt", "ExpiresAt")

	if diff := cmp.Diff(want, got, opts); diff != "" {
		t.Error(diff)
	}
}
