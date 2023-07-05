package data

import (
	"fmt"
	"strings"
	"time"

	valid "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/mlafeldt/listkeeper/functions/internal/twitter"
)

const (
	userIndex = "UserIndex"

	typeUser          = "User"
	typeFollowerList  = "FollowerList"
	typeFollowerEvent = "FollowerEvent"

	FollowerStateNew              = "NEW"
	FollowerStateLost             = "LOST"
	FollowerStateReasonFollowed   = "FOLLOWED"
	FollowerStateReasonUnfollowed = "UNFOLLOWED"
	FollowerStateReasonDeleted    = "DELETED"
	FollowerStateReasonSuspended  = "SUSPENDED"
)

type User struct {
	ID              string      `json:"id" dynamo:"UserID"`
	Handle          string      `json:"handle"`
	Name            string      `json:"name"`
	Location        string      `json:"location,omitempty"`
	Bio             string      `json:"bio,omitempty"`
	ProfileImageURL string      `json:"profileImageUrl"`
	AccessToken     string      `json:"-"`
	AccessSecret    string      `json:"-"`
	Slack           SlackConfig `json:"slack"`
	IgnoreFollowers []string    `json:"ignoreFollowers,omitempty" dynamo:",set,omitempty"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
	LastLogin       time.Time   `json:"lastLogin"`
	LastIP          string      `json:"-"`
	LoginsCount     int64       `json:"-"`
	IDP             string      `json:"-" dynamo:"IdP"`
}

type SlackConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhookUrl,omitempty" dynamo:",omitempty"`
	Channel    string `json:"channel,omitempty" dynamo:",omitempty"`
}

func (c SlackConfig) Validate() error {
	return valid.ValidateStruct(&c,
		valid.Field(&c.Enabled),
		valid.Field(&c.WebhookURL, valid.When(c.Enabled, valid.Required, is.URL)),
		valid.Field(&c.Channel), // FIXME: too permissive
	)
}

type userItem struct {
	PK        string
	SK        string
	UserIndex string
	Type      string

	*User
}

func NewUser(id string) *User {
	now := time.Now()
	return &User{ID: id, CreatedAt: now, UpdatedAt: now}
}

func (u *User) IgnoresFollower(id, handle string) bool {
	for _, ignore := range u.IgnoreFollowers {
		if ignore == id {
			return true
		}
		if strings.TrimPrefix(ignore, "@") == handle {
			return true
		}
	}
	return false
}

func (u *User) Validate() error {
	err := valid.ValidateStruct(u,
		valid.Field(&u.ID, valid.Required),
		valid.Field(&u.Handle, valid.Required),
		valid.Field(&u.Name, valid.Required),
		valid.Field(&u.Location),
		valid.Field(&u.Bio),
		valid.Field(&u.ProfileImageURL, valid.Required, is.URL),
		valid.Field(&u.AccessToken, valid.Required),
		valid.Field(&u.AccessSecret, valid.Required),
		valid.Field(&u.Slack),
		valid.Field(&u.IgnoreFollowers, valid.Each(valid.Required)), // FIXME: too permissive
		valid.Field(&u.CreatedAt, valid.Required),
		valid.Field(&u.UpdatedAt, valid.Required, valid.Min(u.CreatedAt)),
		valid.Field(&u.LastLogin, valid.Required),
		valid.Field(&u.LastIP, valid.Required, is.IP),
		valid.Field(&u.LoginsCount, valid.Required),
		valid.Field(&u.IDP),
	)
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s -> %s", typeUser, err) //nolint:errorlint
}

func (u *User) pk() string { return "USER#" + u.ID }
func (u *User) sk() string { return "USER#" + u.ID }

func (u *User) toItem() *userItem {
	return &userItem{
		PK:        u.pk(),
		SK:        u.sk(),
		UserIndex: u.pk(),
		Type:      typeUser,
		User:      u,
	}
}

type FollowerList struct {
	UserID         string
	S3Bucket       string
	S3Key          string
	TotalFollowers int
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

type followerListItem struct {
	PK   string
	SK   string
	TTL  time.Time `dynamo:",unixtime"`
	Type string

	*FollowerList
}

func (l *FollowerList) Validate() error {
	err := valid.ValidateStruct(l,
		valid.Field(&l.UserID, valid.Required),
		valid.Field(&l.S3Bucket, valid.Required),
		valid.Field(&l.S3Key, valid.Required),
		valid.Field(&l.CreatedAt, valid.Required),
		valid.Field(&l.ExpiresAt, valid.Required, valid.Min(l.CreatedAt.Add(1*time.Hour))),
	)
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s -> %s", typeFollowerList, err) //nolint:errorlint
}

func (l *FollowerList) pk() string { return "USER#" + l.UserID }
func (l *FollowerList) sk() string { return "FOLLOWERS#" + l.CreatedAt.Format(time.RFC3339) }

func (l *FollowerList) toItem() *followerListItem {
	return &followerListItem{
		PK:           l.pk(),
		SK:           l.sk(),
		TTL:          l.ExpiresAt,
		Type:         typeFollowerList,
		FollowerList: l,
	}
}

type FollowerEvent struct {
	ID                  string        `json:"id" dynamo:"EventID"`
	UserID              string        `json:"userId" tstype:"-"` // FIXME: required by notify-user
	TotalFollowers      int           `json:"totalFollowers"`
	Follower            *twitter.User `json:"follower" tstype:",required"`
	FollowerState       string        `json:"followerState" tstype:"'NEW' | 'LOST'"`
	FollowerStateReason string        `json:"followerStateReason" tstype:"'FOLLOWED' | 'UNFOLLOWED' | 'DELETED' | 'SUSPENDED'"`
	CreatedAt           time.Time     `json:"createdAt"`
	ExpiresAt           time.Time     `json:"-"`
}

type followerEventItem struct {
	PK   string
	SK   string
	TTL  time.Time `dynamo:",unixtime"`
	Type string

	*FollowerEvent
}

func (e *FollowerEvent) Validate() error {
	err := valid.ValidateStruct(e,
		valid.Field(&e.ID, valid.Required),
		valid.Field(&e.UserID, valid.Required),
		valid.Field(&e.TotalFollowers),
		valid.Field(&e.Follower, valid.Required),
		valid.Field(&e.FollowerState, valid.Required),
		valid.Field(&e.FollowerStateReason, valid.Required),
		valid.Field(&e.CreatedAt, valid.Required),
		valid.Field(&e.ExpiresAt, valid.Required, valid.Min(e.CreatedAt.Add(1*time.Hour))),
	)
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s -> %s", typeFollowerEvent, err) //nolint:errorlint
}

func (e *FollowerEvent) pk() string { return "USER#" + e.UserID }
func (e *FollowerEvent) sk() string { return "EVENT#" + e.ID }

func (e *FollowerEvent) toItem() *followerEventItem {
	return &followerEventItem{
		PK:            e.pk(),
		SK:            e.sk(),
		TTL:           e.ExpiresAt,
		Type:          typeFollowerEvent,
		FollowerEvent: e,
	}
}

type UserSignupEvent struct {
	UserID string `tstype:"-"`
}
