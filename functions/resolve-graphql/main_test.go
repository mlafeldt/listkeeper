package main

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var compareErrors = cmp.Comparer(func(x, y error) bool {
	if x == nil || y == nil {
		return x == nil && y == nil
	}
	return x.Error() == y.Error()
})

func TestUserID(t *testing.T) {
	tests := []struct {
		event   appSyncEvent
		argName string
		userID  string
		err     error
	}{
		{
			event: appSyncEvent{},
			err:   errors.New("unauthorized: user ID must not be empty"),
		},
		{
			event: appSyncEvent{
				Arguments: map[string]interface{}{"id": "1234"},
			},
			argName: "id",
			userID:  "1234",
		},
		{
			event: appSyncEvent{
				Arguments: map[string]interface{}{"userId": "1234"},
			},
			argName: "userId",
			userID:  "1234",
		},
		{
			event: appSyncEvent{
				Arguments: map[string]interface{}{"id": "twitter|1234"},
			},
			argName: "id",
			userID:  "1234",
		},
		{
			event: appSyncEvent{
				Arguments: map[string]interface{}{"id": "twitter|1234"},
				Identity:  Identity{Sub: "twitter|1234"},
			},
			argName: "id",
			userID:  "1234",
		},
		{
			event: appSyncEvent{
				Arguments: map[string]interface{}{"id": "twitter|1234"},
				Identity:  Identity{Sub: "twitter|5678"},
			},
			argName: "id",
			err:     errors.New("unauthorized: user ID must match subject claim"),
		},
	}

	for _, test := range tests {
		id, err := test.event.userID(test.argName)

		if diff := cmp.Diff(test.err, err, compareErrors); diff != "" {
			t.Fatal(diff)
		}

		if diff := cmp.Diff(test.userID, id); diff != "" {
			t.Error(diff)
		}
	}
}
