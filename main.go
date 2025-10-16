package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
)

func main() {
	fmt.Println("Captured Clipboard Demo")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Determine hidden file path in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("could not determine home directory: %v", err)
	}
	hiddenFile := filepath.Join(homeDir, ".clipboard_capture")

	// start clipboard reading loop
	readClipboard(ticker, hiddenFile)
}

func readClipboard(ticker *time.Ticker, hiddenFile string) {
	var last string

	for {
		txt, err := clipboard.ReadAll()
		if err != nil {
			// If clipboard is locked by another app, skip until next tick
			log.Printf("warning: could not read clipboard: %v", err)
		} else if txt != last {
			fmt.Printf("[%s]: %q\n", time.Now().Format(time.RFC3339), summarize(txt))
			last = txt
			if last != "" {
				storeClipboard(hiddenFile, last)
			}
		}
		<-ticker.C
	}
}

func storeClipboard(path, data string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("error: could not open clipboard file: %v", err)
		return
	}
	defer f.Close()
	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s]: %s\n", timestamp, data)
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

