package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestS3Upload(t *testing.T) {
	data, _ := ioutil.ReadFile("event-demo.txt")
	s3event := events.S3Event{}
	json.Unmarshal(data, &s3event)
	LambdaHandler(context.TODO(), s3event)
}
