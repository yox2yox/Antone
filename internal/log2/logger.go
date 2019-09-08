package log2

import (
	"log"
	"os"
)

var Debug *log.Logger
var Err *log.Logger
var debugfile *os.File
var errfile *os.File

func init() {
	file, err := os.OpenFile(`../log/debug.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	errfile, err = os.OpenFile(`../log/error.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	Debug = log.New(file, "[DEBUG]", log.LstdFlags)
	debugfile = file
	Err = log.New(errfile, "[ERROR]", log.LstdFlags)
}

func Close() {
	debugfile.Close()
	errfile.Close()
}
