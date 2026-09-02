package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func initLogger() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}
