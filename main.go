package main

import (
	"fmt"
	"log"
	"time"
	"os"
	"path/filepath"

	"github.com/atotto/clipboard"
)

func main() {
	log.Println("Captured Clipboard Demo")
	var last string

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Determine hidden file path in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("could not determine home directory: %v", err)
	}
	hiddenFile := filepath.Join(homeDir, ".clipboard_capture")

	// initial read (non-blocking)
	if txt, err := clipboard.ReadAll(); err == nil {
		last = txt
		fmt.Printf("[%s] initial clipboard: %q\n", time.Now().Format(time.RFC3339), summarize(last))
		storeClipboard(hiddenFile, last)
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
			storeClipboard(hiddenFile, txt)
		}
	}
}

// storeClipboard appends clipboard data to a hidden file with restricted permissions
func storeClipboard(path, data string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("error: could not open clipboard file: %v", err)
		return
	}
	defer f.Close()
	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s]\n%s\n---\n", timestamp, data)
	if _, err := f.WriteString(entry); err != nil {
		log.Printf("error: could not write to clipboard file: %v", err)
	}
}

func summarize(s string) string {
	// Keep output readable: show up to 300 chars
	if len(s) > 300 {
		return s[:300] + "...(truncated)"
	}
	return s
}
