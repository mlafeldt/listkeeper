package data

import (
	"context"
)

//nolint:gofumpt
type TableAPI interface {
	CreateUser(ctx context.Context, u *User) error
	UpdateUser(ctx context.Context, u *User) error
	RegisterUser(ctx context.Context, u *User) error
	GetUser(ctx context.Context, userID string) (*User, error)
	DeleteUser(ctx context.Context, userID string) error
	NewUserIter() UserIter

	CreateFollowerList(ctx context.Context, l *FollowerList) error
	GetUserAndLatestFollowerLists(ctx context.Context, userID string, limit int64) (*User, []*FollowerList, error)
	GetLatestFollowerLists(ctx context.Context, userID string, limit int64) ([]*FollowerList, error)

	CreateFollowerEvent(ctx context.Context, e *FollowerEvent) error
	GetLatestFollowerEvents(ctx context.Context, userID string, limit int64) ([]*FollowerEvent, error)
}

type UserIter interface {
	Next(ctx context.Context) *User
	Err() error
}

var (
	_ TableAPI = (*Table)(nil)
	_ UserIter = (*userIter)(nil)
)
