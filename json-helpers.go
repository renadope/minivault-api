package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Envelope map[string]any

var ErrJSONMarshal = errors.New("error marshaling JSON")

func WriteJSON(responseWriter http.ResponseWriter, status int, data Envelope, headers http.Header) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ErrJSONMarshal
	}
	jsonData = append(jsonData, '\n')
	for key, value := range headers {
		responseWriter.Header()[key] = value
	}
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(status)
	_, err = responseWriter.Write(jsonData)
	if err != nil {
		return err
	}
	return nil

}

func ReadJSON(responseWriter http.ResponseWriter, httpRequest *http.Request, dst any) error {
	maxBytes := 1_048_576
	httpRequest.Body = http.MaxBytesReader(responseWriter, httpRequest.Body, int64(maxBytes))
	dec := json.NewDecoder(httpRequest.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("body contains badly formed JSON")
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON at character:%d", syntaxError.Offset)
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field == "" {
				return fmt.Errorf("body contains incorrect JSON type for field: %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type at character %d", unmarshalTypeError.Offset)

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key: %s", fieldName)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return fmt.Errorf("unaccounted for error:%w", err)
		}
	}
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must contain a single JSON object")
	}
	return nil
}
