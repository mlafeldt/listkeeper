package main

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/request"
	lambdasvc "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/google/go-cmp/cmp"

	"github.com/mlafeldt/listkeeper/functions/internal/data"
)

type tableStub struct {
	data.TableAPI

	users []*data.User
}

func (t *tableStub) NewUserIter() data.UserIter {
	return &userIterStub{users: t.users}
}

type userIterStub struct {
	data.UserIter

	users []*data.User
	index int
}

func (iter *userIterStub) Next(context.Context) *data.User {
	if iter.index >= len(iter.users) {
		return nil
	}
	u := iter.users[iter.index]
	iter.index++
	return u
}

func (iter *userIterStub) Err() error {
	return nil
}

type lambdaStub struct {
	lambdaiface.LambdaAPI
}

func (*lambdaStub) InvokeWithContext(_ context.Context, _ *lambdasvc.InvokeInput, _ ...request.Option) (*lambdasvc.InvokeOutput, error) {
	return nil, nil //nolint:nilnil
}

func TestEnqueueNoUsers(t *testing.T) {
	h := handler{
		table: &tableStub{},
	}

	want := &output{
		TotalUsers: 0,
	}

	got, err := h.handle(context.Background(), events.CloudWatchEvent{})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func TestEnqueueOneUser(t *testing.T) {
	h := handler{
		table: &tableStub{
			users: []*data.User{
				data.NewUser("111"),
			},
		},
		lambda: &lambdaStub{},
	}

	want := &output{
		UserIDs:    []string{"111"},
		TotalUsers: 1,
	}

	got, err := h.handle(context.Background(), events.CloudWatchEvent{})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func TestEnqueueSomeUsers(t *testing.T) {
	h := handler{
		table: &tableStub{
			users: []*data.User{
				data.NewUser("111"),
				data.NewUser("222"),
				data.NewUser("333"),
			},
		},
		lambda: &lambdaStub{},
	}

	want := &output{
		UserIDs:    []string{"111", "222", "333"},
		TotalUsers: 3,
	}

	got, err := h.handle(context.Background(), events.CloudWatchEvent{})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}
