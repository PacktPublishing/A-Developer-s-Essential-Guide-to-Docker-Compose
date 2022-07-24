package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	SQS_TOPIC_ENV               = "SQS_TOPIC"
	AWS_REGION_ENV              = "AWS_DEFAULT_REGION"
	SQS_ENDPOINT_ENV            = "SQS_ENDPOINT"
	S3STORE_LAMBDA_ENDPOINT_ENV = "S3STORE_LAMBDA_ENDPOINT"
)

func sqsSession() (*sqs.SQS, error) {
	session, _ := session.NewSession()

	return sqs.New(session, aws.NewConfig().WithEndpoint(os.Getenv(SQS_ENDPOINT_ENV)).WithRegion(os.Getenv(AWS_REGION_ENV))), nil
}

func main() {
	session, _ := sqsSession()

	for {

		queueUrl := aws.String(os.Getenv(SQS_TOPIC_ENV))
		if msgResult, err := session.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl: queueUrl,
		}); err == nil {

			if msgResult != nil && len(msgResult.Messages) > 0 {
				sqsEvent := map[string][]*sqs.Message{
					"Records": msgResult.Messages,
				}

				log.Printf("Dispatching %v received messages", len(msgResult.Messages))

				marshalled, _ := json.Marshal(sqsEvent)
				http.Post(os.Getenv(S3STORE_LAMBDA_ENDPOINT_ENV), "application/json", bytes.NewBuffer(marshalled))

				for i := 0; i < len(msgResult.Messages); i++ {

					session.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      queueUrl,
						ReceiptHandle: msgResult.Messages[i].ReceiptHandle,
					})

				}
			}
		}
		time.Sleep(1 * time.Second)
	}

}
