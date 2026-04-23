package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	type streams struct {
		Streams []struct {
			Width       int    `json:"width"`
			Height      int    `json:"height"`
			AspectRatio string `json:"display_aspect_ratio"`
		} `json:"streams"`
	}

	command := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	stdOutBuf := bytes.Buffer{}
	command.Stdout = &stdOutBuf

	err := command.Run()
	if err != nil {
		return "", err
	}

	var fileRatio streams
	if err := json.Unmarshal(stdOutBuf.Bytes(), &fileRatio); err != nil {
		return "", err
	}

	switch fileRatio.Streams[0].AspectRatio {
	case "16:9":
		return "landscape", nil
	case "9:16":
		return "portrait", nil
	default:
		return "other", nil
	}
}

func processVideoForFastStart(filePath string) (string, error) {
	outputFile := filePath + ".processing"
	command := exec.Command("ffmpeg", "-i",
		filePath,
		"-c",
		"copy",
		"-movflags",
		"faststart",
		"-f",
		"mp4",
		outputFile,
	)

	err := command.Run()
	if err != nil {
		return "", err
	}

	return outputFile, nil
}
