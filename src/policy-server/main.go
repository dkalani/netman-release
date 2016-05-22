package main

import (
	"encoding/json"
	"flag"
	"lib/marshal"
	"os"
	"policy-server/config"
	"policy-server/handlers"
	"policy-server/store"

	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
	"github.com/tedsuo/rata"
)

func main() {
	logger := lager.NewLogger("policy-server")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))
	logger.Info("starting-setup")
	defer logger.Info("stopping")

	var configFilePath string
	const configFileFlag = "configFile"

	flag.StringVar(&configFilePath, configFileFlag, "", "")
	flag.Parse()
	logger.Info("flag-parse-complete")

	conf, err := config.ParseConfigFile(configFilePath)
	if err != nil {
		logger.Error("config", err)
		os.Exit(1)
	}

	marshaler := marshal.MarshalFunc(json.Marshal)
	unmarshaler := marshal.UnmarshalFunc(json.Unmarshal)

	packetTagger, err := store.NewMemoryTagger(4)
	if err != nil {
		logger.Error("packet tag", err)
		os.Exit(1)
	}

	rulesStore := store.NewMemoryStore(packetTagger)

	rataHandlers := rata.Handlers{}
	rataHandlers["rules_list"] = &handlers.RulesList{
		Logger:    logger,
		Marshaler: marshaler,
		Store:     rulesStore,
	}
	rataHandlers["rules_add"] = &handlers.RulesAdd{
		Logger:      logger,
		Unmarshaler: unmarshaler,
		Store:       rulesStore,
	}
	rataHandlers["rules_delete"] = &handlers.RulesDelete{
		Logger:      logger,
		Unmarshaler: unmarshaler,
		Store:       rulesStore,
	}
	rataHandlers["whitelists"] = &handlers.Whitelists{
		Logger:    logger,
		Marshaler: marshaler,
		Store:     rulesStore,
	}

	routes := rata.Routes{
		{Name: "rules_list", Method: "GET", Path: "/rules"},
		{Name: "rules_add", Method: "POST", Path: "/rules/add"},
		{Name: "rules_delete", Method: "POST", Path: "/rules/delete"},
		{Name: "whitelists", Method: "GET", Path: "/whitelists"},
	}

	rataRouter, err := rata.NewRouter(routes, rataHandlers)
	if err != nil {
		logger.Fatal("create-rata-route", err)
	}

	httpServer := http_server.New(conf.ListenAddress, rataRouter)

	members := grouper.Members{
		{"http_server", httpServer},
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	logger.Info("ifrit-invoke")
	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("ifrit-wait")
	err = <-monitor.Wait()
	if err != nil {
		logger.Fatal("ifrit-wait", err)
	}
}
