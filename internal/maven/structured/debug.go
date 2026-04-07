package structured

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	if os.Getenv("MAVEN_PARSER_DEBUG") != "" {
		logger = log.New(os.Stderr, "[maven-parser-debug] ", log.LstdFlags)
	} else {
		logger = log.New(os.Stderr, "", 0)
		logger.SetOutput(os.Stdout) // (silence logger; replaced below)
		logger.SetOutput(os.Stderr) // by default, output to Stderr (optional)
		// For production silence, use: logger.SetOutput(io.Discard)

	}
}

func Debugf(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf(format, v...)
	}
}
