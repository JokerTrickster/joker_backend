package logger

import (
	"log"
	"os"
)

var (
	infoLog  *log.Logger
	errorLog *log.Logger
)

func Init(level string) {
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(message string) {
	if infoLog != nil {
		infoLog.Println(message)
	}
}

func Error(message string) {
	if errorLog != nil {
		errorLog.Println(message)
	}
}
