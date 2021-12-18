package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kuai6/urlator/pkg/renderer"
	"github.com/kuai6/urlator/pkg/runner"
	"github.com/rs/zerolog/log"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal().Msg("no urls specified")
	}

	list := os.Args[1:]

	rnr := runner.NewRunner()

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		cancel()
	}()

	defer func() {
		cancel()
		signal.Stop(c)
	}()

	res, err := rnr.Run(ctx, list)
	if err != nil {
		cancel()
		log.Fatal().Err(err).Msg("runner error")
	}

	rdr := renderer.NewRenderer()

	fmt.Print(rdr.Render(res))

	ctx.Done()
}
