package evb

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
)

type API interface {
	Send(ctx context.Context, eventType string, events ...interface{}) error
}

var _ API = (*Client)(nil)

type Config struct {
	EventBusName    string
	EventSourceName string
}

type Client struct {
	config *Config
	svc    eventbridgeiface.EventBridgeAPI
}

func NewClient(p client.ConfigProvider, cfgs ...*Config) *Client {
	config := &Config{}

	for _, cfg := range cfgs {
		if cfg.EventBusName != "" {
			config.EventBusName = cfg.EventBusName
		}
		if cfg.EventSourceName != "" {
			config.EventSourceName = cfg.EventSourceName
		}
	}

	return &Client{
		config: config,
		svc:    eventbridge.New(p),
	}
}

func (c *Client) Send(ctx context.Context, eventType string, events ...interface{}) error {
	if len(events) == 0 {
		return nil
	}

	entries := make([]*eventbridge.PutEventsRequestEntry, len(events))

	for i, event := range events {
		detail, err := json.Marshal(event)
		if err != nil {
			return err
		}

		entries[i] = &eventbridge.PutEventsRequestEntry{
			EventBusName: aws.String(c.config.EventBusName),
			Source:       aws.String(c.config.EventSourceName),
			DetailType:   aws.String(eventType),
			Detail:       aws.String(string(detail)),
		}
	}

	_, err := c.svc.PutEventsWithContext(ctx, &eventbridge.PutEventsInput{
		Entries: entries,
	})

	return err
}
