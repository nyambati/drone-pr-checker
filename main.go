package main

import (
	"log"

	"github.com/nyambati/drone-pr-checker/internal/config"
	"github.com/nyambati/drone-pr-checker/internal/github"
	"github.com/nyambati/drone-pr-checker/internal/plugin"
)

func main() {
	config, err := config.New()

	if err != nil {
		log.Fatal(err)
	}

	plugin := plugin.New(
		config.Settings,
		github.New(config.Github.Token),
	)
	plugin.Report()
}
