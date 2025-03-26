package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// InitLogger initializes the logging configuration.
func InitLogger(logFile string) error {
	// Ensure the directory exists
	dir := filepath.Dir(logFile)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Open the log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Set up logging to both file and stdout
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return nil
}

// LogInfo logs informational messages.
func LogInfo(message string) {
	log.Println("INFO: " + message)
}