package log2

import (
	"log"
	"os"
)

var Debug *log.Logger
var Err *log.Logger
var TestER *log.Logger
var debugfile *os.File
var errfile *os.File
var testErrRatefile *os.File
var Export *log.Logger
var exportResult *os.File

func init() {
	file, err := os.OpenFile("log/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		Debug = log.New(os.Stdout, "[DEBUG]", log.LstdFlags)
	} else {
		Debug = log.New(file, "[DEBUG]", log.LstdFlags)
		debugfile = file
	}
	errfile, err = os.OpenFile("log/error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		Err = log.New(os.Stderr, "[ERROR]", log.LstdFlags)
	} else {
		Err = log.New(errfile, "[ERROR]", log.LstdFlags)
	}
	testErrRatefile, err = os.OpenFile("log/testER.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		TestER = log.New(os.Stderr, "[TEST]", log.LstdFlags)
	} else {
		TestER = log.New(testErrRatefile, "[TEST]", log.LstdFlags)
	}
	exportResult, err = os.OpenFile("log/expResult.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		Export = log.New(os.Stdout, "", 0)
	} else {
		Export = log.New(exportResult, "", 0)
	}
}

func Close() {
	if debugfile != nil {
		debugfile.Close()
	}
	if errfile != nil {
		errfile.Close()
	}
	if testErrRatefile != nil {
		testErrRatefile.Close()
	}
	if exportResult != nil {
		exportResult.Close()
	}
}
