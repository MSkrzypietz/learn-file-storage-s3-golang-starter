package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
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
