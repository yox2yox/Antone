package log2

import (
	"log"
	"os"
)

var Debug *log.Logger
var Err *log.Logger

func init() {
	Debug = log.New(os.Stdout, "[DEBUG]", log.LstdFlags)
	Err = log.New(os.Stderr, "[ERROR]", log.LstdFlags)
}
