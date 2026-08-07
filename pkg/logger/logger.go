package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

var (
	logFile      *os.File
	serialCounter uint64
)

// InitLogger initializes logging to both stdout and logs/app.log file
func InitLogger(logPath string) error {
	dir := filepath.Dir(logPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	logFile = f

	// MultiWriter routes log output to stdout and log file simultaneously
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Printf("[INIT] Logger initialized. Logging errors to terminal and file: %s\n", logPath)
	return nil
}

// Close closes the underlying log file on server shutdown
func Close() {
	if logFile != nil {
		_ = logFile.Close()
	}
}

// GenerateCode builds a unique error code based on folder name, file name, function name, and serial counter.
// Format: UKJ001
// U - 1st letter of Folder
// K - 1st letter of File
// J - 1st letter of Function
// 001 - 3-digit serial number
func GenerateCode(folder, file, fnName string) string {
	folderChar := 'X'
	if len(folder) > 0 {
		folderChar = rune(strings.ToUpper(string(folder[0]))[0])
	}

	// Remove extension from filename if present
	baseFile := filepath.Base(file)
	baseFile = strings.TrimSuffix(baseFile, filepath.Ext(baseFile))
	fileChar := 'X'
	if len(baseFile) > 0 {
		fileChar = rune(strings.ToUpper(string(baseFile[0]))[0])
	}

	// Extract simple function name if struct prefix exists (e.g. (*UserService).CreateUser -> CreateUser)
	cleanFn := fnName
	if idx := strings.LastIndex(fnName, "."); idx != -1 {
		cleanFn = fnName[idx+1:]
	}
	cleanFn = strings.TrimPrefix(cleanFn, "*")

	fnChar := 'X'
	if len(cleanFn) > 0 {
		fnChar = rune(strings.ToUpper(string(cleanFn[0]))[0])
	}

	num := atomic.AddUint64(&serialCounter, 1)
	serial := num % 1000
	if serial == 0 {
		serial = 1000
	}

	return fmt.Sprintf("%c%c%c%03d", folderChar, fileChar, fnChar, serial)
}

// LogError logs an error to terminal and app.log with automatic code generation based on runtime caller
func LogError(err error) string {
	if err == nil {
		return ""
	}

	folder := "internal"
	file := "app"
	fnName := "Handler"

	// Inspect call stack to automatically determine folder, file, and function name
	pc, fileP, _, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(fileP)
		dir := filepath.Dir(fileP)
		folder = filepath.Base(dir)

		if fn := runtime.FuncForPC(pc); fn != nil {
			fnName = fn.Name()
		}
	}

	code := GenerateCode(folder, file, fnName)
	log.Printf("[ERROR] [%s] %s | File: %s/%s | Func: %s | Time: %s\n",
		code, err.Error(), folder, file, fnName, time.Now().Format(time.RFC3339))
	return code
}

// LogWithCode logs an explicit error with specified folder, file, fnName, and error message
func LogWithCode(folder, file, fnName string, err error) string {
	if err == nil {
		return ""
	}
	code := GenerateCode(folder, file, fnName)
	log.Printf("[ERROR] [%s] %s | File: %s/%s | Func: %s | Time: %s\n",
		code, err.Error(), folder, file, fnName, time.Now().Format(time.RFC3339))
	return code
}

// LogInfo logs informational messages to terminal and app.log
func LogInfo(msg string) {
	log.Printf("[INFO] %s\n", msg)
}
