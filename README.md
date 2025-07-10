# MiniVault API

A lightweight local REST API that generates AI responses using Ollama.
Built for the ModelVault (Minivault) take-home project, this API provides both standard JSON responses and real-time streaming capabilities along with logging.

## Prerequisites

- **Go 1.23+** installed
- **[Ollama](https://ollama.ai/download)** installed and running
- At least one language model downloaded (recommended: `ollama pull llama3.2:1b`)

## Setup & Running

1. **Clone or download the project**
   ```bash
   cd minivault-api
   ```

2. **Ensure Ollama is running**
   ```bash
   ollama serve
   ```
   Starting Ollama may vary by operating system

3. **Start the API server**
   ```bash
   go run .
   ```
   if the dependencies don't automatically install, you can run:
   
   ```bash
   go mod download
   go run .
   ```

   The server will start on `http://localhost:4000` by default.

4. **Test the API**
   ```bash
   curl -X POST http://localhost:4000/generate \
     -H "Content-Type: application/json" \
     -d '{"prompt": "Give me the recipe for chocolate chip cookies"}'
   ```
   ```bash
   curl -N -X POST http://localhost:4000/generate-stream \
     -H "Content-Type: application/json" \
     -d '{"prompt": "Give me the recipe for chocolate chip cookies with a surprise twist"}'
   ```

## API Endpoints

### `POST /generate`
Returns a complete JSON response after generation finishes.

**Request:**
```json
{ "prompt": "Your prompt here" }
```

**Response:**
```json
{ "response": "Generated response here" }
```

### `POST /generate-stream`
Returns a streaming text response with tokens appearing in real-time.

**Request:**
```json
{ "prompt": "Your prompt here" }
```

**Response:** Plain text stream (Content-Type: text/plain)


## Configuration

The API supports command-line configuration:

```bash
go run . -port=8080 -llm-url="http://localhost:11434" -llm-name="mistral:7b"
```

**Available flags:**
- `-port`: API server port (default: 4000)
- `-llm-url`: Ollama base URL (default: http://localhost:11434)
- `-llm-name`: Model name to use (default: llama3.2:1b)

## Testing

### Using curl
```bash
# JSON endpoint
curl -X POST http://localhost:4000/generate \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Write a haiku about coding"}'

# Streaming endpoint  
curl -N -X POST http://localhost:4000/generate-stream \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Tell me about space"}'
```

### Using Postman
Import the `postman_collection.json` file into Postman for easy API testing with pre-configured requests.

## Logging
If you don't see the `logs/log.jsonl`, it will be created on the initial ```go run . ``` command. 

All API interactions are automatically logged to `logs/log.jsonl` in JSON Lines format:

```json
{"timestamp":"2025-07-09T16:31:07-04:00","prompt":"Hello world","response":"Generated response here"}
```

## Implementation Notes

### Design Choices
- **Graceful degradation**: If Ollama is unavailable, the API returns stubbed error responses
- **Dual interfaces**: Both JSON and streaming endpoints to demonstrate different response patterns
- **Structured logging**: Using Go's `slog` for consistent, structured logs
- **Configuration flexibility**: Command-line flags allow easy deployment in different environments

### Error Handling
- Input validation (empty prompts return 400 errors)
- LLM connection failures trigger fallback responses
- Logging failures don't break API responses
- Proper HTTP status codes for different error types




