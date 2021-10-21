package main

import (
	"io"
	"os"

	"github.com/sentientwaffle/cl/internal/colorize"
)

func main() {
	c := colorize.NewColorizer(os.Stdin)
	for {
		b, err := c.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Stderr.WriteString("\n")
			os.Exit(1)
		}
		os.Stdout.Write(b)
	}
}
