package main

/*
 */
import "C"

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// export run_service
func run_service(host *C.char, port uint16) uint16 {
	return 0
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func main() {}
