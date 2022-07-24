package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{

		SharedConfigState: session.SharedConfigEnable,
	}))

	s3 := s3.New(sess, aws.NewConfig().WithEndpoint("http://localhost:9090").WithRegion("us-west-2"))

	buckets, _ := s3.ListBuckets(nil)

	for i, bucket := range buckets.Buckets {

		fmt.Printf("%d: %s\n", i, *bucket.Name)

	}
}
