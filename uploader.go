package main

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
)

var sess = connectAWS()

func connectAWS() *session.Session {
	conf := aws.Config{
		Region:           aws.String("us-east-1"),
		Credentials:      credentials.NewSharedCredentials("", "ci"),
		Endpoint:         aws.String("http://ci.labs:9000"),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess, err := session.NewSession(&conf)
	if err != nil {
		panic(err)
	}
	return sess
}

func UploadBytesS3(key string, bucket string, b []byte) error {
	_, err := s3.New(sess).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(bucket),
		Key:                aws.String(key),
		ACL:                aws.String("private"),
		Body:               bytes.NewReader(b),
		ContentLength:      aws.Int64(int64(len(b))),
		ContentDisposition: aws.String("attachment"),
	})
	return err
}

func UploadS3(filename string, bucket string, key string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	err = UploadBytesS3(key, bucket, buffer)

	if err != nil {
		Log.Printf("Failed to upload: %s", err.Error())
	}
	return err
}

func CreateBucket(name string) error {
	_, err := s3.New(sess).CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				Log.Printf("Bucket {%s} already exists and is owned by you, not an issue", name)
				return nil
			case s3.ErrCodeBucketAlreadyExists:
				Log.Printf("Bucket {%s} already exists but it is owned by someone else, pick a new name", name)
				return aerr
			default:
				Log.Printf("An unknown issue popped up during bucket creation: %s", aerr.Error())
				return aerr
			}
		}
	}
	return nil
}
