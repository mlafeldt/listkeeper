package data

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/go-cmp/cmp"
	"github.com/guregu/dynamo"

	"github.com/mlafeldt/listkeeper/functions/internal/twitter"
)

var created, _ = time.Parse(time.RFC822, "07 Nov 20 21:04 UTC")

var compareErrors = cmp.Comparer(func(x, y error) bool {
	if x == nil || y == nil {
		return x == nil && y == nil
	}
	return x.Error() == y.Error()
})

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		user *User
		err  error
	}{
		{
			user: &User{},
			err:  errors.New("User -> AccessSecret: cannot be blank; AccessToken: cannot be blank; LastIP: cannot be blank; LoginsCount: cannot be blank; createdAt: cannot be blank; handle: cannot be blank; id: cannot be blank; lastLogin: cannot be blank; name: cannot be blank; profileImageUrl: cannot be blank; updatedAt: cannot be blank."), //nolint:revive
		},
		{
			user: &User{
				ID:              "1234",
				Handle:          "alice",
				Name:            "Alice",
				ProfileImageURL: "https://example.com/profile.png",
				AccessToken:     "token",
				AccessSecret:    "secret",
				CreatedAt:       created,
				UpdatedAt:       created.Add(1 * time.Hour),
				LastLogin:       created.Add(1 * time.Hour),
				LastIP:          "1.2.3.4",
				LoginsCount:     3,
			},
			err: nil,
		},
	}

	for _, test := range tests {
		err := test.user.Validate()

		if diff := cmp.Diff(test.err, err, compareErrors); diff != "" {
			t.Error(diff)
		}
	}
}

func TestUser_ToItem(t *testing.T) {
	u := User{
		ID:              "1234",
		Handle:          "alice",
		Name:            "Alice",
		Location:        "Wonderland",
		ProfileImageURL: "https://example.com/profile.png",
		AccessToken:     "token",
		AccessSecret:    "secret",
		CreatedAt:       created,
		UpdatedAt:       created.Add(1 * time.Hour),
		LastLogin:       created.Add(1 * time.Hour),
		LastIP:          "1.2.3.4",
		LoginsCount:     3,
	}

	want := map[string]*dynamodb.AttributeValue{
		"PK":              {S: aws.String("USER#1234")},
		"SK":              {S: aws.String("USER#1234")},
		"UserIndex":       {S: aws.String("USER#1234")},
		"UserID":          {S: aws.String("1234")},
		"Type":            {S: aws.String("User")},
		"Handle":          {S: aws.String("alice")},
		"Name":            {S: aws.String("Alice")},
		"Location":        {S: aws.String("Wonderland")},
		"ProfileImageURL": {S: aws.String("https://example.com/profile.png")},
		"AccessToken":     {S: aws.String("token")},
		"AccessSecret":    {S: aws.String("secret")},
		"Slack": {M: map[string]*dynamodb.AttributeValue{
			"Enabled": {BOOL: aws.Bool(false)},
		}},
		"CreatedAt":   {S: aws.String("2020-11-07T21:04:00Z")},
		"UpdatedAt":   {S: aws.String("2020-11-07T22:04:00Z")},
		"LastLogin":   {S: aws.String("2020-11-07T22:04:00Z")},
		"LastIP":      {S: aws.String("1.2.3.4")},
		"LoginsCount": {N: aws.String("3")},
	}

	if err := u.Validate(); err != nil {
		t.Fatal(err)
	}

	got, err := dynamo.MarshalItem(u.toItem())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func TestFollowerList_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		list *FollowerList
		err  error
	}{
		{
			list: &FollowerList{
				S3Bucket: "some-bucket",
				S3Key:    "/some/path",
			},
			err: errors.New("FollowerList -> CreatedAt: cannot be blank; ExpiresAt: cannot be blank; UserID: cannot be blank."), //nolint:revive
		},
		{
			list: &FollowerList{
				UserID:    "1234",
				CreatedAt: now,
				ExpiresAt: now.Add(1 * time.Hour),
			},
			err: errors.New("FollowerList -> S3Bucket: cannot be blank; S3Key: cannot be blank."), // nolint:revive
		},
		{
			list: &FollowerList{
				UserID:    "1234",
				S3Bucket:  "some-bucket",
				S3Key:     "/some/path",
				CreatedAt: now,
				ExpiresAt: now.Add(24 * time.Hour),
			},
			err: nil,
		},
	}

	for _, test := range tests {
		err := test.list.Validate()

		if diff := cmp.Diff(test.err, err, compareErrors); diff != "" {
			t.Error(diff)
		}
	}
}

func TestFollowerList_ToItem(t *testing.T) {
	l := FollowerList{
		UserID:         "1234",
		S3Bucket:       "some-bucket",
		S3Key:          "/some/path",
		TotalFollowers: 1000,
		CreatedAt:      created,
		ExpiresAt:      created.Add(24 * time.Hour),
	}

	want := map[string]*dynamodb.AttributeValue{
		"PK":             {S: aws.String("USER#1234")},
		"SK":             {S: aws.String("FOLLOWERS#2020-11-07T21:04:00Z")},
		"TTL":            {N: aws.String("1604869440")},
		"Type":           {S: aws.String("FollowerList")},
		"UserID":         {S: aws.String("1234")},
		"S3Bucket":       {S: aws.String("some-bucket")},
		"S3Key":          {S: aws.String("/some/path")},
		"TotalFollowers": {N: aws.String("1000")},
		"CreatedAt":      {S: aws.String("2020-11-07T21:04:00Z")},
		"ExpiresAt":      {S: aws.String("2020-11-08T21:04:00Z")},
	}

	if err := l.Validate(); err != nil {
		t.Fatal(err)
	}

	got, err := dynamo.MarshalItem(l.toItem())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func TestFollowerEvent_Validate(t *testing.T) {
	tests := []struct {
		event *FollowerEvent
		err   error
	}{
		{
			event: &FollowerEvent{},
			err:   errors.New("FollowerEvent -> ExpiresAt: cannot be blank; createdAt: cannot be blank; follower: cannot be blank; followerState: cannot be blank; followerStateReason: cannot be blank; id: cannot be blank; userId: cannot be blank."), //nolint:revive
		},
		{
			event: &FollowerEvent{
				ID:                  "some-event-id",
				UserID:              "some-user-id",
				Follower:            &twitter.User{},
				FollowerState:       "NEW",
				FollowerStateReason: "FOLLOWED",
				CreatedAt:           created,
				ExpiresAt:           created.Add(24 * time.Hour),
			},
			err: nil,
		},
	}

	for _, test := range tests {
		err := test.event.Validate()

		if diff := cmp.Diff(test.err, err, compareErrors); diff != "" {
			t.Error(diff)
		}
	}
}

func TestFollowerEvent_ToItem(t *testing.T) {
	e := &FollowerEvent{
		ID:             "some-event-id",
		UserID:         "some-user-id",
		TotalFollowers: 200,
		Follower: &twitter.User{
			ID:             "123",
			Handle:         "alice",
			Name:           "Alice",
			Location:       "Wonderland",
			Bio:            "I ❤️  adventures",
			TotalFollowers: 100,
		},
		FollowerState:       "NEW",
		FollowerStateReason: "FOLLOWED",
		CreatedAt:           created,
		ExpiresAt:           created.Add(24 * time.Hour),
	}

	want := map[string]*dynamodb.AttributeValue{
		"PK":             {S: aws.String("USER#some-user-id")},
		"SK":             {S: aws.String("EVENT#some-event-id")},
		"TTL":            {N: aws.String("1604869440")},
		"Type":           {S: aws.String("FollowerEvent")},
		"EventID":        {S: aws.String("some-event-id")},
		"UserID":         {S: aws.String("some-user-id")},
		"TotalFollowers": {N: aws.String("200")},
		"Follower": {M: map[string]*dynamodb.AttributeValue{
			"ID":             {S: aws.String("123")},
			"Handle":         {S: aws.String("alice")},
			"Name":           {S: aws.String("Alice")},
			"Location":       {S: aws.String("Wonderland")},
			"Bio":            {S: aws.String("I ❤️  adventures")},
			"Protected":      {BOOL: aws.Bool(false)},
			"TotalFollowers": {N: aws.String("100")},
		}},
		"FollowerState":       {S: aws.String("NEW")},
		"FollowerStateReason": {S: aws.String("FOLLOWED")},
		"CreatedAt":           {S: aws.String("2020-11-07T21:04:00Z")},
		"ExpiresAt":           {S: aws.String("2020-11-08T21:04:00Z")},
	}

	if err := e.Validate(); err != nil {
		t.Fatal(err)
	}

	got, err := dynamo.MarshalItem(e.toItem())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}
