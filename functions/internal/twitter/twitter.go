package twitter

import (
	"context"
	"errors"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	twitterOAuth1 "github.com/dghubble/oauth1/twitter"
)

type API interface {
	FollowerIDs(ctx context.Context, accessToken, accessSecret string) ([]int64, error)
	CurrentUser(ctx context.Context, accessToken, accessSecret string) (*User, error)
	UserByID(ctx context.Context, accessToken, accessSecret string, userID int64) (*User, error)
}

var _ API = (*Client)(nil)

type Client struct {
	config *oauth1.Config
}

func NewClient(consumerKey, consumerSecret string) *Client {
	return &Client{
		config: &oauth1.Config{
			ConsumerKey:    consumerKey,
			ConsumerSecret: consumerSecret,
			Endpoint:       twitterOAuth1.AuthorizeEndpoint,
		},
	}
}

func (c *Client) newClientWithContext(ctx context.Context, accessToken, accessSecret string) *twitter.Client {
	token := oauth1.NewToken(accessToken, accessSecret)
	return twitter.NewClient(c.config.Client(ctx, token))
}

// Due to Twitter's API rate limiting, this function will only return up to
// 75,000 followers (15 requests * 5000 items, over 15 minutes).
func (c *Client) FollowerIDs(ctx context.Context, accessToken, accessSecret string) ([]int64, error) {
	const (
		maxRequests  = 15
		maxBatchSize = 5000
	)

	var (
		tc  = c.newClientWithContext(ctx, accessToken, accessSecret)
		ids = []int64{}
	)

	for req, cursor := 0, int64(-1); req < maxRequests && cursor != 0; req++ {
		params := twitter.FollowerIDParams{Cursor: cursor, Count: maxBatchSize}
		followers, _, err := tc.Followers.IDs(&params)
		if err != nil {
			return nil, makeErr(err)
		}
		ids = append(ids, followers.IDs...)
		cursor = followers.NextCursor
	}

	return ids, nil
}

type User struct {
	ID              string `json:"id"`
	Handle          string `json:"handle,omitempty"`
	Name            string `json:"name,omitempty"`
	Location        string `json:"location,omitempty"`
	Bio             string `json:"bio,omitempty"`
	ProfileImageURL string `json:"profileImageUrl,omitempty"`
	Protected       bool   `json:"protected"`
	TotalFollowers  int    `json:"totalFollowers"`
}

func (c *Client) CurrentUser(ctx context.Context, accessToken, accessSecret string) (*User, error) {
	tc := c.newClientWithContext(ctx, accessToken, accessSecret)

	u, _, err := tc.Accounts.VerifyCredentials(nil)
	if err != nil {
		return nil, makeErr(err)
	}

	return makeUser(u), nil
}

func (c *Client) UserByID(ctx context.Context, accessToken, accessSecret string, userID int64) (*User, error) {
	tc := c.newClientWithContext(ctx, accessToken, accessSecret)

	u, _, err := tc.Users.Show(&twitter.UserShowParams{UserID: userID})
	if err != nil {
		return nil, makeErr(err)
	}

	return makeUser(u), nil
}

func makeUser(u *twitter.User) *User {
	return &User{
		ID:              u.IDStr,
		Handle:          u.ScreenName,
		Name:            u.Name,
		Location:        u.Location,
		Bio:             u.Description,
		ProfileImageURL: u.ProfileImageURLHttps,
		Protected:       u.Protected,
		TotalFollowers:  u.FollowersCount,
	}
}

//nolint:gomnd
func makeErr(err error) error {
	var apiErr twitter.APIError

	if errors.As(err, &apiErr) {
		if !apiErr.Empty() {
			switch apiErr.Errors[0].Code {
			case 50:
				return ErrUserNotFound
			case 63:
				return ErrUserSuspended
			case 88:
				return ErrRateLimitExceeded
			case 89:
				return ErrInvalidToken
			}
		}
	}

	return err
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserSuspended     = errors.New("user suspended")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrInvalidToken      = errors.New("invalid or expired token")
)
