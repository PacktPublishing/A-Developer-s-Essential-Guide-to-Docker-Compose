package main

import (
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

func main() {
	sess, _ := session.NewSession(&aws.Config{

		Region: aws.String("us-west-2"),

		Credentials: credentials.NewStaticCredentials("fakeMyKeyId", "fakeSecretAccessKey", ""),
	})

	svc := dynamodb.New(sess, aws.NewConfig().WithEndpoint("http://localhost:8000").WithRegion("eu-west-2"))

	item := Subscribe{

		Email: "john@doe.com",

		Topic: "what I subscribed",
	}

	av, _ := dynamodbattribute.MarshalMap(item)

	input := &dynamodb.PutItemInput{

		Item: av,

		TableName: aws.String("Newsletter"),
	}

	svc.PutItem(input)
}
