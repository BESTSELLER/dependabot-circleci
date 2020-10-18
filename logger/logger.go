package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// Init sets the global loglevel
func Init() {
	// default is info
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	loglevel := os.Getenv("DEPENDABOT_LOGLEVEL")
	if loglevel == "" {
		return
	}
	if strings.ToLower(loglevel) == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
