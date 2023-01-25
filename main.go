package main

import (
	"flag"
	"strings"

	"github.com/grafana/grafana-deployment-modes/pkg/modules"
)

func main() {
	targetPtr := flag.String("target", "all", "a space separated list of modules to run. default is all")
	flag.Parse()

	targets := []string{"all"}

	// if target is not empty, split it by space delimiter to get a list of targets
	if targetPtr != nil && *targetPtr != "" {
		targets = strings.Split(*targetPtr, " ")
	}

	// register modules
	modules, err := modules.New(targets)
	if err != nil {
		panic(err)
	}

	// use service manager to start targets and dependencies
	if err := modules.Run(); err != nil {
		panic(err)
	}

}
