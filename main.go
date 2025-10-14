package main

import (
	"fmt"
	"log"
	"time"

	"github.com/atotto/clipboard"
)

func main() {
	log.Println("Clipboard demo started. Press Ctrl+C to stop.")
	var last string

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// initial read (non-blocking)
	if txt, err := clipboard.ReadAll(); err == nil {
		last = txt
		fmt.Printf("[%s] initial clipboard: %q\n", time.Now().Format(time.RFC3339), summarize(last))
	}

	for range ticker.C {
		txt, err := clipboard.ReadAll()
		if err != nil {
			// If clipboard is locked by another app, skip this tick
			log.Printf("warning: could not read clipboard: %v", err)
			continue
		}
		if txt != last {
			fmt.Printf("[%s] clipboard changed: %q\n", time.Now().Format(time.RFC3339), summarize(txt))
			last = txt
		}
	}
}

func summarize(s string) string {
	// Keep output readable: show up to 300 chars
	if len(s) > 300 {
		return s[:300] + "...(truncated)"
	}
	return s
}
