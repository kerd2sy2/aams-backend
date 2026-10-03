package safe

import (
	"log"
	"runtime/debug"
)

// Go runs a background goroutine protected with panic recovery to prevent crashing the server process.
func Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC RECOVERED] in safe goroutine: %v\nStack Trace:\n%s", r, string(debug.Stack()))
			}
		}()
		fn()
	}()
}
