package netapi

import (
	"fmt"
	"policy-server/models"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/pivotal-cf-experimental/rainmaker"
)

const (
	CommandAllow    = "net-allow"
	CommandDisallow = "net-disallow"
	CommandList     = "net-list"
)

type client interface {
	AddRule(rule models.Rule) error
	DeleteRule(rule models.Rule) error
	ListRules() ([]models.Rule, error)
}

type userLogger interface {
	Printf(format string, v ...interface{})
}

type Runner struct {
	Client        client
	UserLogger    userLogger
	CliConnection plugin.CliConnection
	Rainmaker     rainmaker.Client
}

func (r *Runner) getRule(sourceName, destinationName string) (models.Rule, error) {
	app1, err := r.CliConnection.GetApp(sourceName)
	if err != nil {
		return models.Rule{}, fmt.Errorf("getting app %s: %s", sourceName, err)
	}
	app2, err := r.CliConnection.GetApp(destinationName)
	if err != nil {
		return models.Rule{}, fmt.Errorf("getting app %s: %s", destinationName, err)
	}
	return models.Rule{Source: app1.Guid, Destination: app2.Guid}, nil
}

func (r *Runner) resolveAndPrettyPrint(rule models.Rule, token string) (string, error) {
	token = strings.TrimPrefix(token, "bearer ") // rainmaker adds its own bearer
	app1, err := r.Rainmaker.Applications.Get(rule.Source, token)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %s", rule.Source, err)
	}
	app2, err := r.Rainmaker.Applications.Get(rule.Destination, token)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %s", rule.Destination, err)
	}

	sourceName := app1.Name
	destinationName := app2.Name
	return fmt.Sprintf("%s --> %s", sourceName, destinationName), nil
}

func (r *Runner) Run(args []string) error {
	command := args[0]

	isLoggedIn, err := r.CliConnection.IsLoggedIn()
	if err != nil {
		return fmt.Errorf("checking logged in: %s", err)
	}
	if !isLoggedIn {
		return fmt.Errorf("please log in")
	}

	token, err := r.CliConnection.AccessToken()
	if err != nil {
		return fmt.Errorf("getting token: %s", err)
	}

	switch command {
	case CommandList:
		rules, err := r.Client.ListRules()
		if err != nil {
			return fmt.Errorf("list: %s", err)
		}
		prettyPrintedRules := []string{}
		for _, rule := range rules {
			prettyPrintedRule, err := r.resolveAndPrettyPrint(rule, token)
			if err != nil {
				return fmt.Errorf("parsing rules: %s", err)
			}
			prettyPrintedRules = append(prettyPrintedRules, prettyPrintedRule)
		}
		r.UserLogger.Printf("net-allow rules:")
		for _, ppr := range prettyPrintedRules {
			r.UserLogger.Printf("%s\n", ppr)
		}
	case CommandAllow, CommandDisallow:
		if len(args) != 3 {
			return fmt.Errorf("missing required arguments, try -h")
		}
		sourceName := args[1]
		destinationName := args[2]
		rule, err := r.getRule(sourceName, destinationName)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		switch command {
		case CommandAllow:
			err = r.Client.AddRule(rule)
			if err != nil {
				return fmt.Errorf("allow: %s", err)
			}
			r.UserLogger.Printf("allowed %s --> %s\n", sourceName, destinationName)
		case CommandDisallow:
			err = r.Client.DeleteRule(rule)
			if err != nil {
				return fmt.Errorf("disallow: %s", err)
			}
			r.UserLogger.Printf("disallowed %s --> %s\n", sourceName, destinationName)
		}
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
	return nil
}
