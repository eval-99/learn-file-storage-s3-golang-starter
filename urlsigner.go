package main

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	obj := s3.GetObjectInput{Bucket: &bucket, Key: &key}
	preSigneds3Client := s3.NewPresignClient(s3Client)
	PreSignedReq, err := preSigneds3Client.PresignGetObject(context.Background(), &obj, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}

	return PreSignedReq.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}

	urlSlice := strings.Split(*video.VideoURL, ",")
	if len(urlSlice) != 2 {
		return video, nil
	}

	bucket := urlSlice[0]
	key := urlSlice[1]

	signedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, time.Hour)
	if err != nil {
		return database.Video{}, err
	}

	video.VideoURL = &signedURL

	return video, nil
}
