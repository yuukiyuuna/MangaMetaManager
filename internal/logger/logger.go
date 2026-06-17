package logger

import (
	"io"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger() {
	logPath := "logs/mmm.log"
	if _, err := os.Stat("/app/data"); err == nil {
		logPath = "/app/data/logs/mmm.log"
	}

	maxSize := 1024
	maxAge := 30

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSize, // megabytes
		MaxBackups: 3,
		MaxAge:     maxAge, // days
		Compress:   true,   // disabled by default
	}

	// Output to both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, lumberjackLogger)
	log.SetOutput(multiWriter)

	log.Printf("Logger initialized: path=%s, max_size=%dMB, max_age=%ddays", logPath, maxSize, maxAge)
}
