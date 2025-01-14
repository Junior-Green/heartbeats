package logger

import (
	"log"
	"os"
)

func Print(v ...any) {
	log.Println(v...)
}

func Printf(format string, v ...any) {
	log.Printf(format, v...)
}
func Debug(v ...any) {
	if isDebugMode() {
		log.Println(v...)
	}
}

func Debugf(format string, v ...any) {
	if isDebugMode() {
		log.Printf(format, v...)
	}
}

func isDebugMode() bool {
	return os.Getenv("DEBUG") == "true"
}
