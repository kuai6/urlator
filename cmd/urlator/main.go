package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	res, err := rnr.Run(context.Background(), list)
	if err != nil {
		log.Fatal(errors.Wrap(err, "runner error"))
	}

	rdr := renderer.NewRenderer()

	fmt.Print(rdr.Render(res))
}
