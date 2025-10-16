package main

import (
	"fmt"
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
	check(err)

	hiddenFile := filepath.Join(homeDir, ".clipboard_capture")
	
	// start clipboard reading loop
	readClipboard(ticker, hiddenFile)
}

func readClipboard(ticker *time.Ticker, hiddenFile string) {
	var last string

	for {
		txt, err := clipboard.ReadAll()
		check(err)

		if txt != last {
			fmt.Printf("[%s]: %q\n", time.Now().Format(time.RFC3339), summarize(txt))
			last = txt
			if last != "" {
				storeClipboard(hiddenFile, last)
			}
		}
		<-ticker.C //wait for next tick
	}
}

func storeClipboard(path, data string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	check(err)

	defer f.Close()
	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s]: %s\n", timestamp, data)
	_, err = f.WriteString(entry)
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func summarize(s string) string {
	// Keep output readable: show up to 300 chars
	if len(s) > 300 {
		return s[:300] + "...(truncated)"
	}
	return s
}

