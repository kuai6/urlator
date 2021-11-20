package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kuai6/urlator/pkg/renderer"
	"github.com/kuai6/urlator/pkg/runner"
	"github.com/pkg/errors"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("no urls specified")
	}

	list := os.Args[1:]

	rnr := runner.NewRunner()

	ctx, cancel := context.WithCancel(context.Background())
	res, err := rnr.Run(ctx, list)
	if err != nil {
		cancel()
		log.Fatal(errors.Wrap(err, "runner error"))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	defer func() {
		cancel()
		signal.Stop(c)
	}()

	go func() {
		<-c
		cancel()
	}()

	rdr := renderer.NewRenderer()

	fmt.Print(rdr.Render(res))

	ctx.Done()
}
