package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

type LLMConfig struct {
	BaseURL string
	Model   string
	Timeout time.Duration
}
type application struct {
	logger    *slog.Logger
	port      int
	wg        *sync.WaitGroup
	llmConfig *LLMConfig
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var port int
	var baseLLMUrl string
	var model string
	var wg sync.WaitGroup

	flag.IntVar(&port, "port", 4000, "Api server port")
	flag.StringVar(&baseLLMUrl, "llm-url", "http://localhost:11434", "Base URL for the LLM service (e.g., http://localhost:11434 for local Ollama instance)")
	flag.StringVar(&model, "llm-model", "llama3.2:1b", "The name of the LLM Model you would like to use")
	flag.Parse()

	llmConfig := LLMConfig{
		BaseURL: baseLLMUrl,
		Model:   model,
		Timeout: 30 * time.Second,
	}

	app := &application{
		port:      port,
		logger:    logger,
		wg:        &wg,
		llmConfig: &llmConfig,
	}

	err := ensureLogDirectoryExists()
	if err != nil {
		app.logger.Error("error creating log directory", slog.String("error", err.Error()))
		os.Exit(1)
	}

	serverConfig := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}
	app.logger.Info("starting server", "addr", serverConfig.Addr)
	err = Serve(serverConfig, &wg, app.logger)
	if err != nil {
		app.logger.Error("error starting server", "error", err.Error())
		os.Exit(1)
	}
}

func ensureLogDirectoryExists() error {
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		return fmt.Errorf("error creating log directory: %w", err)
	}
	logFile := "logs/log.jsonl"
	_, err = os.Stat(logFile)
	if errors.Is(err, fs.ErrNotExist) {
		file, err := os.Create(logFile)
		if err != nil {
			return fmt.Errorf("error creating log file: %w", err)
		}
		err = file.Close()
		if err != nil {
			return fmt.Errorf("error closing file after creation: %w", err)
		}
	}
	return nil

}

//Improvement notes:
//Logging to a server and not local filesystem, we can use rabbitmq/kafka to do this and handle it in a background job
//More detailed logs, with error in the logs whether it failed at Prompt, or Response generation or in general any other related errors
//Add request IDs for tracing
