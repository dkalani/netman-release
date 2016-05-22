package main

import (
	"cf-cli-plugin/netapi"

	"github.com/cloudfoundry/cli/plugin"
)

func main() {
	plugin.Start(&netapi.Plugin{})
}
