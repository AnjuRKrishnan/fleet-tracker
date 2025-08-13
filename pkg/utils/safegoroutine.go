package utils

import (
	"log"
	"runtime/debug"
)

// SafeGo runs a function in a new goroutine and recovers from any panics.
func SafeGo(f func(), args ...interface{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered in goroutine: %v, args: %v\n%s", r, args, debug.Stack())
			}
		}()
		f()
	}()
}
