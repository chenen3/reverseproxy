package main

import (
	"log"
	"os"
)

type _logger struct {
	Infof  func(format string, v ...interface{})
	Errorf func(format string, v ...interface{})
	Error  func(v ...interface{})
	Fatal  func(v ...interface{})
}

var (
	logger = new(_logger)
	// for access log only
	accessLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
)

func init() {
	logger.Infof = log.New(os.Stderr, "INFO  ", log.Ldate|log.Ltime|log.Lshortfile).Printf
	errLogger := log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Errorf = errLogger.Printf
	logger.Error = errLogger.Print
	logger.Fatal = log.New(os.Stderr, "FATAL ", log.Ldate|log.Ltime|log.Lshortfile).Print
}
