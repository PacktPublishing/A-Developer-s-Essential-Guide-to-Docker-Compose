package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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
			return "Could not add item",err
		}

		return fmt.Sprintf("You have been subscribed to the %s newsletter", subscribe.Topic), nil
	}
}

func dynamoDBSession() (*dynamodb.DynamoDB, error) {
	session, err := createSession()

	if err != nil {
		return nil, err
	}

	return dynamodb.New(session, aws.NewConfig().WithEndpoint("http://host.docker.internal:8000").WithRegion("eu-west-2")), nil
}

func createSession() (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials("fakeMyKeyId", "fakeSecretAccessKey", ""),
	})
}

func main() {
	lambda.Start(HandleRequest)
}
