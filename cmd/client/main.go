package main

import (
	"context"
	"flag"
	"log"

	"github.com/sc-chat/test-chat/internal/sigctx"
	"github.com/sc-chat/test-chat/pkg/client"
)

var (
	addr  string
	name  string
	debug bool
	ms    int
)

func init() {
	flag.StringVar(&addr, "a", "0.0.0.0:8000", "server address")
	flag.StringVar(&name, "n", "", "client name")
	flag.BoolVar(&debug, "d", false, "debug mode")

	flag.Parse()
}

func main() {
	c, err := client.NewClient(addr, name, debug)
	if err != nil {
		log.Fatal(err)
	}

	ctx := sigctx.NewSignalContext(context.Background())

	err = c.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
