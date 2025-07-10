package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Prompt    string `json:"prompt"`
	Response  string `json:"response"`
}

func (app *application) logInteraction(prompt string, response string) error {

	logEntry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Prompt:    prompt,
		Response:  response,
	}

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("error marshalling log entry: %w", err)
	}

	file, err := os.OpenFile("logs/log.jsonl", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			app.logger.Error("error closing log file", "error", err.Error())
		}
	}(file)
	_, err = file.Write(append(jsonData, '\n'))
	if err != nil {
		return fmt.Errorf("error writing to log file: %w", err)
	}
	return nil

}
