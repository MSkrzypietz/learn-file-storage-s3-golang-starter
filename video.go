package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"os/exec"
	"strings"
	"time"
)

type ffprobeResult struct {
	Streams []struct {
		Width  int `json:"width,omitempty"`
		Height int `json:"height,omitempty"`
	} `json:"streams"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	buffer := bytes.NewBuffer(nil)
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = buffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	decoder := json.NewDecoder(buffer)
	cmdResult := ffprobeResult{}
	err = decoder.Decode(&cmdResult)
	if err != nil {
		return "", err
	}

	stream := cmdResult.Streams[0]
	ratio := float64(stream.Width) / float64(stream.Height)
	if ratio > 1.7 && ratio < 1.8 {
		return "16:9", nil
	}

	if ratio > 0.5 && ratio < 0.6 {
		return "9:16", nil
	}

	return "other", nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outputPath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return outputPath, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	s3Parts := strings.Split(*video.VideoURL, ",")
	bucket := s3Parts[0]
	key := s3Parts[1]
	presignedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, 1*time.Minute)
	if err != nil {
		return database.Video{}, err
	}
	video.VideoURL = &presignedURL
	return video, nil
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	client := s3.NewPresignClient(s3Client)
	req, err := client.PresignGetObject(context.Background(), &s3.GetObjectInput{Bucket: &bucket, Key: &key}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
