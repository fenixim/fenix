package utils

import (
	"io"
	"log"
	"os"
	"regexp"
)

var (
	InfoLogger *log.Logger
	WarningLogger *log.Logger
	ErrorLogger *log.Logger
)

type LogLevel int

type colourCodeRemover struct {
    r io.Writer
}

func (c colourCodeRemover) Write(p []byte) (int, error) {
	re, err := regexp.Compile(`\x1b\[[0-9;]*m`)
	if err != nil {
		panic(err)
	}

    return c.r.Write([]byte(re.ReplaceAllLiteral(p, []byte(""))))
}


func InitLogger(level LogLevel, logfile ...string) {
	lfile := ""
	if len(logfile) != 0 {
		lfile = logfile[0]
	}
	file := io.Discard
	if lfile != "" {
		var err error
		file, err = os.OpenFile(lfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
	cc := colourCodeRemover{r: file}
	mw := io.MultiWriter(os.Stdout, cc)
	
	
	var infoLoggerWriter io.Writer = file
	var warningLoggerWriter io.Writer = file
	var errorLoggerWriter io.Writer  = file
	if level >= 3 {
		infoLoggerWriter = mw
	}

	if level >= 2 {
		warningLoggerWriter = mw
	}

	if level >= 1 {
		errorLoggerWriter = mw
	}
	InfoLogger = log.New(infoLoggerWriter, "\033[2;36m[INFO]\033[0m  ", log.Lshortfile | log.Ltime)
	WarningLogger = log.New(warningLoggerWriter, "\033[33m[WARN]\033[0m  ", log.Lshortfile | log.Ltime)
	ErrorLogger = log.New(errorLoggerWriter, "\033[1;31m[ERROR]\033[0m  ", log.Lshortfile | log.Ltime)
}