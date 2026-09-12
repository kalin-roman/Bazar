// Package logger centralizes how the server logs errors, so the
// actual logging implementation (currently stdlib log, writing to
// stderr) can change in one place later without touching every
// handler that reports an error.
package logger

import "log"

// Error logs err tagged with component (e.g. "categoryhttp: List"),
// for an error that's being turned into an opaque 500 response —
// the caller only sees "something went wrong," this is what actually
// shows up in the server's own logs (e.g. Render's) for diagnosing it.
func Error(component string, err error) {
	log.Printf("%s: %v", component, err)
}
