package main

import (
	"fmt"
	"os"

	"github.com/nerdneilsfield/RSSHub-Gateway/cmd"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	if err := cmd.Execute(version, buildTime, gitCommit); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
