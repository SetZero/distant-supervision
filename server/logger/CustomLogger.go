package logger

import (
	"io"
	"log"
	"os"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func init() {
	const output = "console"
	var errorOutput io.Writer
	var warningOutput io.Writer
	var infoOutput io.Writer

	if output == "file" {
		file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		infoOutput = file
		warningOutput = file
		errorOutput =  file
	} else {
		infoOutput = os.Stdout
		warningOutput = os.Stdout
		errorOutput =  os.Stderr
	}

	InfoLogger = log.New(infoOutput, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(warningOutput, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errorOutput, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}