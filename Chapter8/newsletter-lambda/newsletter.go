package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Subscribe struct {
	Email string `json:"email"`
	Topic string `json:"topic"`
}

const TableName = "newsletter"

func HandleRequest(ctx context.Context, subscribe Subscribe) (string, error) {

	log.Printf(subscribe.Email + " " + subscribe.Topic)

	if dynamoDb, err := dynamoDBSession(); err != nil {
		return "Could not process request", err
	} else {
		marshalled, err := dynamodbattribute.MarshalMap(subscribe)
		if err != nil {
			return "could not marshall", err
		}

		input := &dynamodb.PutItemInput{
			Item:      marshalled,
			TableName: aws.String(TableName),
		}

		_, err = dynamoDb.PutItem(input)
		if err != nil {
			return "Could not add item", err
		}

		sendToSQS(subscribe)

		return fmt.Sprintf("You have been subscribed to the %s newsletter", subscribe.Topic), nil
	}
}

func isSimulated() bool {
	if value := os.Getenv("SIMULATED"); len(value) == 0 {
		return false
	} else if value != "true" {
		return false
	}

	return true
}

func dynamoDBSession() (*dynamodb.DynamoDB, error) {
	session, _ := session.NewSession()

	if isSimulated() {
		return dynamodb.New(session, aws.NewConfig().WithEndpoint(os.Getenv("DYNAMODB_ENDPOINT")).WithRegion("eu-west-2")), nil
	} else {
		return dynamodb.New(session), nil
	}
}

func sendToSQS(subscribe Subscribe) {
	if !isSimulated() {
		return
	}

	if session, err := sqsSession(); err == nil {
		if bytes, err := jsonutil.BuildJSON(subscribe); err == nil {
			smsInput := &sqs.SendMessageInput{
				MessageBody: aws.String(string(bytes)),
				QueueUrl:    aws.String(os.Getenv("SQS_TOPIC")),
			}

			if _, err := session.SendMessage(smsInput); err != nil {
				fmt.Println(err)
			}

		} else {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println(err.Error())
	}
}

func sqsSession() (*sqs.SQS, error) {
	session, _ := session.NewSession()

	return sqs.New(session, aws.NewConfig().WithEndpoint(os.Getenv("SQS_ENDPOINT")).WithRegion("eu-west-2")), nil
}
func main() {
	lambda.Start(HandleRequest)
}
