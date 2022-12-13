package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var sess *session.Session
var rateLimit int

func init() {
	rate := os.Getenv("RATE")
	var err error
	rateLimit, err = strconv.Atoi(rate)
	if err != nil {
		rateLimit = 100
	}
}

func LambdaHandler(context context.Context, s3Event events.S3Event) (message string, err error) {
	log.Println(s3Event)
	if len(s3Event.Records) > 1 {
		log.Print("receiver s3 records more than 1 in an event")
	}
	totalFileNumber := 0
	uploadFileNumber := int32(0)
	for _, record := range s3Event.Records {
		sess, _ = session.NewSession(&aws.Config{Region: aws.String(record.AWSRegion)})
		s3record := record.S3
		uploadInfo := fmt.Sprintf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3record.Bucket.Name, s3record.Object.Key)
		log.Print(uploadInfo)
		message += uploadInfo

		prefix := s3record.Object.Key[:strings.LastIndex(s3record.Object.Key, "/")+1]

		downloader := s3manager.NewDownloader(sess)
		file, err := os.Create("/tmp/tmp.zip")
		if err != nil {
			log.Println(err)
			return uploadInfo + err.Error(), err
		}

		_, err = downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(s3record.Bucket.Name),
				Key:    aws.String(s3record.Object.Key),
			})
		message += fmt.Sprintf("%s download zip sucess\n", time.Now())
		file.Close()
		if err != nil {
			log.Println(err)
			return uploadInfo + err.Error(), err
		}

		if err := os.RemoveAll("/tmp/zip"); err != nil {
			log.Println(err)
			return uploadInfo + err.Error(), err
		}

		files, err := unzip("/tmp/tmp.zip", "/tmp/zip")
		if err != nil {
			log.Println(err)
			return uploadInfo + err.Error(), err
		}
		totalFileNumber += len(files)

		log.Printf("upload thread number: %d\n", rateLimit)
		rate := make(chan int, rateLimit)
		var wg sync.WaitGroup
		for _, path := range files {
			rate <- 1
			wg.Add(1)
			go func(bucket string, path string, prefix string) {
				defer wg.Done()
				defer func() {
					<-rate
				}()
				if err := upload(bucket, path, prefix); err == nil {
					atomic.AddInt32(&uploadFileNumber, 1)
				} else {
					log.Println("upload file fail: ", err)
				}
			}(s3record.Bucket.Name, path, prefix)
		}
		wg.Wait()

	}
	message += fmt.Sprintf("%s total file upload: %d, success: %d", time.Now(), totalFileNumber, uploadFileNumber)
	log.Println(message)
	return message, nil
}

func unzip(src, dest string) (files []string, err error) {
	files = make([]string, 0, 10)
	r, err := zip.OpenReader(src)
	if err != nil {
		return files, err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return files, err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			files = append(files, fpath)
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				log.Fatal(err)
				return files, err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return files, err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return files, err
			}
			f.Close()
		}
		rc.Close()
	}
	return files, nil
}

func upload(bucket string, path string, prefix string) error {
	uploader := s3manager.NewUploader(sess)
	file, err := os.Open(path)
	if err != nil {
		log.Println("Failed opening file", path, err)
		return err
	}
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: &bucket,
		Key:    aws.String(prefix + strings.Replace(path, "/tmp/zip/", "", 1)),
		Body:   file,
	})
	if err != nil {
		log.Println("Failed upload file", path, err)
		file.Close()
		return err
	}
	file.Close()
	return nil
}

func main() {
	lambda.Start(LambdaHandler)
}
