package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
	return
}

// newLoggingMiddleware logs the incoming HTTP request & its duration.
func newLoggingMiddleware(ctx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					slog.Error("Request panic",
						"err", err,
						"trace", debug.Stack(),
					)
				}
			}()

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			slog.Log(
				ctx,
				slog.LevelInfo,
				"Request received",
				"status", wrapped.status,
				"method", r.Method,
				"path", r.URL.EscapedPath(),
				"duration", time.Since(start),
			)
		})
	}
}

func newAuthMiddleware(usernameExpected, passwordExpected string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the username and password from the request
			// Authorization header. If no Authentication header is present
			// or the header value is invalid, then the 'ok' return value
			// will be false.
			username, password, ok := r.BasicAuth()
			if ok {
				// Calculate SHA-256 hashes for the provided and expected
				// usernames and passwords.
				usernameHash := sha256.Sum256([]byte(username))
				passwordHash := sha256.Sum256([]byte(password))
				expectedUsernameHash := sha256.Sum256([]byte(usernameExpected))
				expectedPasswordHash := sha256.Sum256([]byte(passwordExpected))

				// Use the subtle.ConstantTimeCompare() function to check if
				// the provided username and password hashes equal the
				// expected username and password hashes. ConstantTimeCompare
				// will return 1 if the values are equal, or 0 otherwise.
				// Importantly, we should to do the work to evaluate both the
				// username and password before checking the return values to
				// avoid leaking information.
				usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
				passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

				// If the username and password are correct, then call
				// the next handler in the chain. Make sure to return
				// afterwards, so that none of the code below is run.
				if usernameMatch && passwordMatch {
					next.ServeHTTP(w, r)
					return
				}
			}

			// If the Authentication header is not present, is invalid, or the
			// username or password is wrong, then set a WWW-Authenticate
			// header to inform the client that we expect them to use basic
			// authentication and send a 401 Unauthorized response.
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}
