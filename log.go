package main

import (
	"fmt"
	"log"
	"os"
)

var (
	_infologger = log.New(os.Stderr, "INFO  ", log.LstdFlags|log.Lshortfile)
	_errLogger  = log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lshortfile)
)

func logInfof(format string, a ...interface{}) {
	_infologger.Output(2, fmt.Sprintf(format, a...))
}

func logError(a ...interface{}) {
	_errLogger.Output(2, fmt.Sprint(a...))
}

func logErrorf(format string, a ...interface{}) {
	_errLogger.Output(2, fmt.Sprintf(format, 2))
}

// avoid flooding log while benchmark
var verbose bool

func setVerbose(b bool) {
	verbose = b
}
