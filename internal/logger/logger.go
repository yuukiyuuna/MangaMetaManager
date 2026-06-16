package logger

import (
	"io"
	"log"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger() {
	logPath := viper.GetString("logging.path")
	if logPath == "" {
		// Check if we are in Docker and have /app/data
		if _, err := os.Stat("/app/data"); err == nil {
			logPath = "/app/data/logs/mmm.log"
		} else {
			logPath = "logs/mmm.log"
		}
	}

	maxSize := viper.GetInt("logging.max_size_mb")
	if maxSize == 0 {
		maxSize = 1024 // 1GB
	}

	maxAge := viper.GetInt("logging.max_age_days")
	if maxAge == 0 {
		maxAge = 30
	}

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
