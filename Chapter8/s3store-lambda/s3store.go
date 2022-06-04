package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	SUBSCRIPTION_BUCKET_ENV = "SUBSCRIPTION_BUCKET"
	SIMULATED_ENV           = "SIMULATED"
	S3_ENDPOINT_ENV         = "S3_ENDPOINT"
	AWS_REGION_ENV          = "AWS_DEFAULT_REGION"
)

type Subscribe struct {
	Email string `json:"email"`
	Topic string `json:"topic"`
}

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	session := s3Session()

	for _, message := range sqsEvent.Records {
		var subscribe Subscribe
		json.Unmarshal([]byte(message.Body), &subscribe)

		key := fmt.Sprintf("%s.%d", hash(subscribe.Email), time.Now().UnixNano()/int64(time.Millisecond))

		marshalled, _ := json.Marshal(subscribe)

		session.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(os.Getenv(SUBSCRIPTION_BUCKET_ENV)),
			Key:    aws.String(key),
			Body:   bytes.NewReader(marshalled),
		})

		fmt.Println("Stored sqs event")
	}

	return nil
}

func hash(email string) string {
	data := []byte(email)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func isSimulated() bool {
	if value := os.Getenv(SIMULATED_ENV); len(value) == 0 {
		return false
	} else if value != "true" {
		return false
	}

	return true
}

func s3Session() *s3.S3 {
	session, _ := session.NewSession()

	if isSimulated() {
		return s3.New(session, aws.NewConfig().WithEndpoint(os.Getenv(S3_ENDPOINT_ENV)).WithRegion(os.Getenv(AWS_REGION_ENV)))
	} else {
		return s3.New(session, aws.NewConfig().WithRegion(os.Getenv(AWS_REGION_ENV)))
	}
}

func main() {
	lambda.Start(HandleRequest)
}
