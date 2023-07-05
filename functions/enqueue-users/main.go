package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	lambdasvc "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/davecgh/go-spew/spew"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"github.com/mlafeldt/listkeeper/functions/internal/data"
)

type output struct {
	UserIDs    []string
	TotalUsers int
}

type handler struct {
	table        data.TableAPI
	lambda       lambdaiface.LambdaAPI
	functionName string
}

func main() {
	var env struct {
		TableName    string `envconfig:"TABLE_NAME" required:"true"`
		FunctionName string `envconfig:"FUNCTION_NAME" required:"true"`
	}
	envconfig.MustProcess("", &env)

	sess := session.Must(session.NewSession())
	h := handler{
		table:        data.NewTable(sess, env.TableName),
		lambda:       lambdasvc.New(sess),
		functionName: env.FunctionName,
	}

	lambda.Start(h.handle)
}

func (h *handler) handle(ctx context.Context, event events.CloudWatchEvent) (*output, error) {
	spew.Printf("event = %+v\n", event)

	var userIDs []string
	iter := h.table.NewUserIter()

	for {
		user := iter.Next(ctx)
		if user == nil {
			break
		}

		_, err := h.lambda.InvokeWithContext(ctx, &lambdasvc.InvokeInput{
			FunctionName:   aws.String(h.functionName),
			Payload:        []byte(fmt.Sprintf(`{"UserID": "%s"}`, user.ID)),
			InvocationType: aws.String(lambdasvc.InvocationTypeEvent),
		})
		if err != nil {
			return nil, errors.Wrapf(err,
				"failed to start function %s for user with ID %s", h.functionName, user.ID)
		}

		userIDs = append(userIDs, user.ID)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	// FIXME: this will break for many users
	out := output{UserIDs: userIDs, TotalUsers: len(userIDs)}

	spew.Printf("output = %+v\n", out)

	return &out, nil
}
