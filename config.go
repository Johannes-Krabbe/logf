package main

import (
	"encoding/json"
	"os"
)

type TransformConfig struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type Config struct {
	TransformLogs bool            `json:"transformLogs"`
	Transform     TransformConfig `json:"transform"`
}

func loadConfig() Config {
	cfg := Config{
		Transform: TransformConfig{
			Timestamp: "timestamp",
			Level:     "level",
			Message:   "",
		},
	}
	data, err := os.ReadFile("logf.json")
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	return cfg
}
