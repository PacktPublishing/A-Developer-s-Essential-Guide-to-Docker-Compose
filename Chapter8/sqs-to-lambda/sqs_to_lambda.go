package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func sqsSession() (*sqs.SQS, error) {
	session, _ := session.NewSession()

	return sqs.New(session, aws.NewConfig().WithEndpoint(os.Getenv("SQS_ENDPOINT")).WithRegion("eu-west-2")), nil
}

func main() {
	session, _ := sqsSession()

	for {

		if msgResult, err := session.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl: aws.String(os.Getenv("SQS_TOPIC")),
		}); err == nil {
			if msgResult != nil {
				for i := 0; i < len(msgResult.Messages); i++ {
					fmt.Println(*msgResult.Messages[i].Body)
				}

			}
		}
		time.Sleep(1 * time.Second)
	}

}
