package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

func (app *application) GenerateHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var input GenerateRequest
	err := ReadJSON(responseWriter, httpRequest, &input)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.Prompt) == "" {
		http.Error(responseWriter, "prompt cannot be empty", http.StatusBadRequest)
		return
	}

	resp, err := app.ollamaFullResponse(input.Prompt)
	if err != nil {
		err = WriteJSON(responseWriter, http.StatusInternalServerError, Envelope{"response": "stubbed error message"}, nil)
		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	err = WriteJSON(responseWriter, http.StatusOK, Envelope{"response": resp}, nil)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	err = app.logInteraction(input.Prompt, resp)
	if err != nil {
		app.logger.Error("error logging interaction", slog.String("error", err.Error()))
		return
	}
}

func (app *application) GenerateStreamHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var input GenerateRequest
	var fullResponse strings.Builder
	err := ReadJSON(responseWriter, httpRequest, &input)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.Prompt) == "" {
		http.Error(responseWriter, "prompt cannot be empty", http.StatusBadRequest)
		return
	}

	responseWriter.Header().Set("Content-Type", "text/plain")
	responseWriter.Header().Set("Cache-Control", "no-cache")
	responseWriter.Header().Set("Connection", "keep-alive")

	flusher, ok := responseWriter.(http.Flusher)
	if !ok {
		http.Error(responseWriter, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	stream, err := app.ollamaStream(input.Prompt)
	if err != nil {
		app.logger.Error("error calling ollama", slog.String("error", err.Error()))

		fallbackMsg := "Error connecting to LLM. Here's a stubbed response to your prompt."
		_, writeErr := fmt.Fprintf(responseWriter, "%s", fallbackMsg)
		if writeErr != nil {
			app.logger.Error("failed to write fallback", slog.String("error", writeErr.Error()))
			return
		}
		flusher.Flush()

		interactionErr := app.logInteraction(input.Prompt, fallbackMsg)
		if interactionErr != nil {
			app.logger.Error("failed to log fallback interaction", "error", interactionErr)
		}
		return
	}
	defer func(stream io.ReadCloser) {
		err := stream.Close()
		if err != nil {
			app.logger.Error("error closing stream", slog.String("error", err.Error()))
		}
	}(stream)

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		var response map[string]interface{}
		err := json.Unmarshal([]byte(line), &response)
		if err != nil {
			app.logger.Error("error parsing ollama response", "error", err)
			continue
		}
		token, ok := response["response"].(string)
		if ok && token != "" {
			_, err := fmt.Fprintf(responseWriter, "%s", token)
			if err != nil {
				app.logger.Error("error writing token", "error", err)
				break
			}
			flusher.Flush()
			fullResponse.WriteString(token)
		}
	}

	err = scanner.Err()
	if err != nil {
		app.logger.Error("scanner error", slog.String("error", err.Error()))
	}

	err = app.logInteraction(input.Prompt, fullResponse.String())
	if err != nil {
		app.logger.Error("error logging interaction", slog.String("error", err.Error()))
		return
	}
}
