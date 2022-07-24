package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	session, _ := session.NewSession(&aws.Config{

		Region: aws.String("us-west-2"),

		Credentials: credentials.NewStaticCredentials("fakeMyKeyId", "fakeSecretAccessKey", ""),
	})

	svc := sqs.New(session, aws.NewConfig().WithEndpoint("http://localhost:9324").WithRegion("eu-west-2"))

	result, _ := svc.ListQueues(nil)

	for i, url := range result.QueueUrls {

		fmt.Printf("%d: %s\n", i, *url)

	}
}
