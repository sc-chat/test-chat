package main

import (
	"context"
	"flag"
	"log"

	"github.com/sc-chat/test-chat/internal/sigctx"
	"github.com/sc-chat/test-chat/pkg/server"
)

var (
	addr  string
	debug bool
)

func init() {
	flag.StringVar(&addr, "a", "0.0.0.0:8000", "server address")
	flag.BoolVar(&debug, "d", false, "debug mode")

	flag.Parse()
}

func main() {
	s, err := server.NewServer(addr, debug)
	if err != nil {
		log.Fatal(err)
	}
	ctx := sigctx.NewSignalContext(context.Background())

	err = s.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
