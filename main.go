package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func main() {
	port := 5632
	if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		port = p
	}

	logger := log.With().
		Str("logger", "telegram-webhook").
		Int("server-port", port).Logger()

	logger.Info().
		Str("msg", fmt.Sprintf("Server started on port %d", port)).
		Send()

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyPretty, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error().
				Str("msg", fmt.Sprintf("error while reading request body: %v", err)).
				Send()
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, "Error: %v", err)
			return
		}
		var bodycompact bytes.Buffer
		err = json.Compact(&bodycompact, bodyPretty)
		if err != nil {
			logger.Error().
				Str("msg", fmt.Sprintf("error while reading request body: %v", err)).
				Send()
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, "Error: %v", err)
			return
		}
		body := bodycompact.Bytes()

		switch r.URL.Path {
		case "/sendMessage":
			switch r.Method {
			case "POST":
				token := r.URL.Query().Get("token")
				url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
				client := &http.Client{}
				req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
				if err != nil {
					logger.Error().
						Str("msg", fmt.Sprintf("error while creating request: %v", err)).
						Send()
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintf(w, "Error: %v", err)
					return
				}

				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					logger.Error().
						Str("msg", fmt.Sprintf("error while sending request: %v", err)).
						Send()
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintf(w, "Error: %v", err)
					return
				}
				defer func(Body io.ReadCloser) {
					_ = Body.Close()
				}(resp.Body)

				w.WriteHeader(resp.StatusCode)
				respBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logger.Error().
						Str("msg", fmt.Sprintf("error while reading response body: %v", err)).
						Send()
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintf(w, "Error: %v", err)
					return
				}
				_, _ = w.Write(respBody)

				logger.Info().
					Str("method", r.Method).
					Str("url", r.URL.Path).
					RawJSON("body", body).
					Int("status", resp.StatusCode).
					RawJSON("response", respBody).
					Send()
			default:
				_, _ = fmt.Fprintf(w, "Path Not Found: %s", r.URL.Path)
			}
		default:
			_, _ = fmt.Fprintf(w, "Path Not Found: %s", r.URL.Path)
		}
	}))

	if err != nil {
		logger.Error().
			Str("msg", fmt.Sprintf("server crashed with error: %v", err)).
			Send()
	} else {
		logger.Info().
			Str("msg", "server stopped").
			Send()
	}
}
