package data

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/dynamo"
)

// Table implements the Table Module pattern: https://www.martinfowler.com/eaaCatalog/tableModule.html
type Table struct {
	inner           dynamo.Table
	consistentReads bool
}

func NewTable(p client.ConfigProvider, name string) *Table {
	return &Table{
		inner:           dynamo.New(p).Table(name),
		consistentReads: false,
	}
}

func NewConsistentTable(p client.ConfigProvider, name string) *Table {
	return &Table{
		inner:           dynamo.New(p).Table(name),
		consistentReads: true,
	}
}

func (t *Table) CreateUser(ctx context.Context, u *User) error {
	if err := u.Validate(); err != nil {
		return err
	}
	return t.inner.Put(u.toItem()).If("attribute_not_exists(PK)").RunWithContext(ctx)
}

func (t *Table) UpdateUser(ctx context.Context, u *User) error {
	if err := u.Validate(); err != nil {
		return err
	}
	u.UpdatedAt = time.Now()
	err := t.inner.Put(u.toItem()).If("attribute_exists(PK)").RunWithContext(ctx)
	if isConditionalCheckErr(err) {
		return ErrUserNotFound
	}
	return err
}

// RegisterUser creates or updates a user, e.g. after login via Auth0.
func (t *Table) RegisterUser(ctx context.Context, u *User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	item := u.toItem()
	err := t.inner.Update("PK", item.PK).Range("SK", item.PK).
		If("'LastLogin' <> ?", item.LastLogin).
		Set(userIndex, item.UserIndex).
		Set("UserID", item.ID).
		Set("Handle", item.Handle).
		Set("Name", item.Name).
		Set("Location", item.Location).
		Set("Bio", item.Bio).
		Set("ProfileImageURL", item.ProfileImageURL).
		Set("AccessToken", item.AccessToken).
		Set("AccessSecret", item.AccessSecret).
		SetIfNotExists("CreatedAt", item.CreatedAt).
		Set("UpdatedAt", item.UpdatedAt).
		Set("LastLogin", item.LastLogin).
		Set("LastIP", item.LastIP).
		Set("LoginsCount", item.LoginsCount).
		Set("IdP", item.IDP).
		Set("Type", typeUser).
		ValueWithContext(ctx, u)

	if isConditionalCheckErr(err) {
		// User wasn't updated, return current data
		u2, err := t.GetUser(ctx, u.ID)
		if err != nil {
			return err
		}
		*u = *u2
		return nil
	}
	return err
}

func (t *Table) GetUser(ctx context.Context, userID string) (*User, error) {
	u := NewUser(userID)
	err := t.inner.Get("PK", u.pk()).
		Index(userIndex).
		Consistent(t.consistentReads).
		OneWithContext(ctx, u)
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	return u, nil
}

func (t *Table) DeleteUser(ctx context.Context, userID string) error {
	u := NewUser(userID)
	err := t.inner.Delete("PK", u.pk()).Range("SK", u.sk()).If("attribute_exists(PK)").RunWithContext(ctx)
	if isConditionalCheckErr(err) {
		return ErrUserNotFound
	}
	return err
}

func (t *Table) NewUserIter() UserIter {
	return &userIter{
		inner: t.inner.Scan().Index(userIndex).Consistent(t.consistentReads).Iter(),
	}
}

type userIter struct {
	inner dynamo.PagingIter
}

func (iter *userIter) Next(ctx context.Context) *User {
	var u User
	if iter.inner.NextWithContext(ctx, &u) {
		return &u
	}
	return nil
}

func (iter *userIter) Err() error {
	return iter.inner.Err()
}

func (t *Table) CreateFollowerList(ctx context.Context, l *FollowerList) error {
	if err := l.Validate(); err != nil {
		return err
	}
	return t.inner.Put(l.toItem()).If("attribute_not_exists(PK)").RunWithContext(ctx)
}

func (t *Table) GetUserAndLatestFollowerLists(ctx context.Context, userID string, limit int64) (*User, []*FollowerList, error) {
	u := NewUser(userID)

	// Since the returned items have mixed types, we need to decode them
	// individually via dynamo.UnmarshalItem after the fact.
	var items []map[string]*dynamodb.AttributeValue

	err := t.inner.Get("PK", u.pk()).
		Limit(limit+1).
		Order(dynamo.Descending).
		Consistent(t.consistentReads).
		AllWithContext(ctx, &items)
	if err != nil {
		return nil, nil, err
	}
	if len(items) == 0 {
		return nil, nil, ErrUserNotFound
	}

	if err := dynamo.UnmarshalItem(items[0], u); err != nil {
		return nil, nil, err
	}
	if err := u.Validate(); err != nil {
		return nil, nil, err
	}

	lists := make([]*FollowerList, len(items)-1)
	for i, item := range items[1:] {
		var l FollowerList
		if err := dynamo.UnmarshalItem(item, &l); err != nil {
			return nil, nil, err
		}
		if err := l.Validate(); err != nil {
			return nil, nil, err
		}
		lists[i] = &l
	}

	return u, lists, nil
}

func (t *Table) GetLatestFollowerLists(ctx context.Context, userID string, limit int64) ([]*FollowerList, error) {
	_, lists, err := t.GetUserAndLatestFollowerLists(ctx, userID, limit)
	return lists, err
}

func (t *Table) CreateFollowerEvent(ctx context.Context, e *FollowerEvent) error {
	if err := e.Validate(); err != nil {
		return err
	}
	return t.inner.Put(e.toItem()).If("attribute_not_exists(PK)").RunWithContext(ctx)
}

func (t *Table) GetLatestFollowerEvents(ctx context.Context, userID string, limit int64) ([]*FollowerEvent, error) {
	e := FollowerEvent{ID: "", UserID: userID}

	var events []*FollowerEvent
	err := t.inner.Get("PK", e.pk()).
		Range("SK", dynamo.BeginsWith, e.sk()).
		Limit(limit).
		Order(dynamo.Descending).
		Consistent(t.consistentReads).
		AllWithContext(ctx, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func isConditionalCheckErr(err error) bool {
	var ae awserr.RequestFailure

	if errors.As(err, &ae) {
		return ae.Code() == "ConditionalCheckFailedException"
	}

	return false
}

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrFollowerListNotFound = errors.New("follower list not found")
)
