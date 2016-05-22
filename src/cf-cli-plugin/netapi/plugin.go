package netapi

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	policyClient "policy-server/client"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/pivotal-cf-experimental/rainmaker"
)

type Plugin struct{}

func (p *Plugin) isValidCommand(command string) bool {
	cmds := p.GetMetadata().Commands
	for _, c := range cmds {
		if c.Name == command {
			return true
		}
	}
	return false
}

func (p *Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	logger := log.New(os.Stdout, "", 0)

	if !p.isValidCommand(args[0]) {
		return // may be CLI-MESSAGE-UNINSTALL, just silently return
	}

	apiEndpoint, err := cliConnection.ApiEndpoint()
	if err != nil {
		logger.Fatalf("unable to discover api endpoint: %s", err)
	}

	skipVerifySSL, err := cliConnection.IsSSLDisabled()
	if err != nil {
		logger.Fatalf("unable to discover status of ssl verification: %s", err)
	}

	var traceWriter io.Writer
	if os.Getenv("CF_TRACE") == "true" {
		traceWriter = os.Stdout
	}

	runner := &Runner{
		Client:        policyClient.NewOuterClient("http://127.0.0.1:5555", http.DefaultClient),
		UserLogger:    logger,
		CliConnection: cliConnection,
		Rainmaker: rainmaker.NewClient(rainmaker.Config{
			Host:          apiEndpoint,
			SkipVerifySSL: skipVerifySSL,
			TraceWriter:   traceWriter,
		}),
	}

	if err := runner.Run(args); err != nil {
		logger.Fatalf("%s", err)
	}
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "connet",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 15,
		},
		Commands: []plugin.Command{
			plugin.Command{
				Name:     CommandAllow,
				HelpText: "Allow direct network traffic from one app to another",
				UsageDetails: plugin.Usage{
					Usage: fmt.Sprintf("cf %s SOURCE_APP DESTINATION_APP", CommandAllow),
				},
			},
			plugin.Command{
				Name:     CommandDisallow,
				HelpText: "Remove an existing net-allow rule",
				UsageDetails: plugin.Usage{
					Usage: fmt.Sprintf("cf %s SOURCE_APP DESTINATION_APP", CommandDisallow),
				},
			},
			plugin.Command{
				Name:     CommandList,
				HelpText: "List all network allow rules",
				UsageDetails: plugin.Usage{
					Usage: fmt.Sprintf("cf %s", CommandList),
				},
			},
		},
	}
}
