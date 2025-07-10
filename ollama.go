package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func (app *application) doOllamaRequest(ollamaRequest *OllamaRequest) (*http.Response, error) {
	jsonBody, err := json.Marshal(ollamaRequest)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, app.llmConfig.BaseURL+"/api/generate", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: app.llmConfig.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling ollama: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned non-200 status code: %d", resp.StatusCode)
	}
	return resp, nil

}
func (app *application) ollamaFullResponse(prompt string) (string, error) {

	ollamaRequest := OllamaRequest{
		Prompt: prompt,
		Model:  app.llmConfig.Model,
		Stream: false,
	}
	resp, err := app.doOllamaRequest(&ollamaRequest)
	if err != nil {
		return "", fmt.Errorf("error performing ollama request: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			app.logger.Error("error closing response body", slog.String("error", err.Error()))
		}
	}(resp.Body)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}
	fullResponse, ok := response["response"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain a string")
	}
	return fullResponse, nil
}

func (app *application) ollamaStream(prompt string) (io.ReadCloser, error) {
	ollamaRequest := OllamaRequest{
		Prompt: prompt,
		Model:  app.llmConfig.Model,
		Stream: true,
	}
	resp, err := app.doOllamaRequest(&ollamaRequest)
	if err != nil {
		return nil, fmt.Errorf("error calling ollama: %w", err)
	}
	return resp.Body, nil

}
