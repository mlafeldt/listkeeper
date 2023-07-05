package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/davecgh/go-spew/spew"
	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/auth0.v5/management"

	"github.com/mlafeldt/listkeeper/functions/internal/data"
	"github.com/mlafeldt/listkeeper/functions/internal/evb"
)

type Info struct {
	FieldName        string
	ParentTypeName   string
	Variables        map[string]interface{}
	SelectionSetList []string
}

type Identity struct {
	Sub    string
	Issuer string
	Claims map[string]interface{}
}

// https://docs.aws.amazon.com/appsync/latest/devguide/resolver-context-reference.html
type appSyncEvent struct {
	Info      Info
	Arguments map[string]interface{}
	Identity  Identity
}

const auth0ProviderPrefix = "twitter|"

func (event appSyncEvent) userID(argName string) (string, error) {
	argID, _ := event.Arguments[argName].(string)

	// With OIDC authorization, subject must match user ID
	if sub := event.Identity.Sub; sub != "" && sub != argID {
		return "", errors.New("unauthorized: user ID must match subject claim")
	}

	// Remove IDP prefix from Auth0 user ID if present
	if id := strings.TrimPrefix(argID, auth0ProviderPrefix); id != "" {
		return id, nil
	}

	return "", errors.New("unauthorized: user ID must not be empty")
}

type handler struct {
	table data.TableAPI
	evb   evb.API
	auth0 *management.Management
}

func main() {
	var env struct {
		TableName       string `envconfig:"TABLE_NAME" required:"true"`
		EventBusName    string `envconfig:"EVENT_BUS_NAME" required:"true"`
		EventSourceName string `envconfig:"EVENT_SOURCE_NAME" required:"true"`
		Auth0           struct {
			Domain       string `envconfig:"AUTH0_DOMAIN" required:"true"`
			ClientID     string `envconfig:"AUTH0_CLIENT_ID" required:"true"`
			ClientSecret string `envconfig:"AUTH0_CLIENT_SECRET" required:"true"`
		}
	}
	envconfig.MustProcess("", &env)

	var (
		opts    = management.WithClientCredentials(env.Auth0.ClientID, env.Auth0.ClientSecret)
		mgmt, _ = management.New(env.Auth0.Domain, opts)
		sess    = session.Must(session.NewSession())
	)

	h := handler{
		table: data.NewTable(sess, env.TableName),
		evb: evb.NewClient(sess, &evb.Config{
			EventBusName:    env.EventBusName,
			EventSourceName: env.EventSourceName,
		}),
		auth0: mgmt,
	}

	lambda.Start(h.handle)
}

func (h *handler) handle(ctx context.Context, event appSyncEvent) (interface{}, error) {
	spew.Printf("event = %+v\n", event)

	switch event.Info.FieldName {
	case "registerUser":
		return h.registerUser(ctx, event)
	case "updateUser":
		return h.updateUser(ctx, event)
	case "deleteUser":
		return h.deleteUser(ctx, event)
	default:
		return nil, fmt.Errorf("unable to resolve field %q", event.Info.FieldName)
	}
}

func (h *handler) registerUser(ctx context.Context, event appSyncEvent) (*data.User, error) {
	userID, err := event.userID("id")
	if err != nil {
		return nil, err
	}

	u0, err := h.auth0.User.Read(auth0ProviderPrefix + userID)
	if err != nil {
		return nil, fmt.Errorf("auth0: %w", err)
	}

	user := data.NewUser(userID)
	user.Handle = u0.GetScreenName()
	user.Name = u0.GetName()
	user.Location = u0.GetLocation()
	user.Bio = u0.GetDescription()
	user.ProfileImageURL = u0.GetPicture()

	if len(u0.Identities) > 0 {
		identity := u0.Identities[0]
		user.AccessToken = identity.GetAccessToken()
		user.AccessSecret = identity.GetAccessTokenSecret()
	}

	user.LastLogin = u0.GetLastLogin()
	user.LastIP = u0.GetLastIP()
	user.LoginsCount = u0.GetLoginsCount()
	user.IDP = event.Identity.Issuer

	if err := h.table.RegisterUser(ctx, user); err != nil {
		return nil, err
	}

	if user.LoginsCount == 1 {
		if err := h.evb.Send(ctx, "New User Signup", data.UserSignupEvent{UserID: userID}); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (h *handler) updateUser(ctx context.Context, event appSyncEvent) (*data.User, error) {
	userID, err := event.userID("id")
	if err != nil {
		return nil, err
	}

	user, err := h.table.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var args struct {
		Input struct {
			Slack           *data.SlackConfig `json:"slack"`
			IgnoreFollowers []string          `json:"ignoreFollowers"`
		} `json:"input"`
	}

	if err := mapstructure.Decode(event.Arguments, &args); err != nil {
		return nil, err
	}
	if v := args.Input.Slack; v != nil {
		user.Slack = *v
	}
	if v := args.Input.IgnoreFollowers; v != nil {
		user.IgnoreFollowers = v
	}

	if err := h.table.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (h *handler) deleteUser(ctx context.Context, event appSyncEvent) (string, error) {
	userID, err := event.userID("id")
	if err != nil {
		return "", err
	}

	if err := h.table.DeleteUser(ctx, userID); err != nil {
		return "", err
	}

	if err := h.auth0.User.Delete(auth0ProviderPrefix + userID); err != nil {
		return "", fmt.Errorf("auth0: %w", err)
	}

	return userID, nil
}
