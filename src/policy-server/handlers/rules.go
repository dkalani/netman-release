package handlers

import (
	"io/ioutil"
	"lib/marshal"
	"net/http"
	"policy-server/models"

	"github.com/pivotal-golang/lager"
)

type store interface {
	Add(logger lager.Logger, rule models.Rule) error
	Delete(logger lager.Logger, rule models.Rule) error
	List(logger lager.Logger) ([]models.Rule, error)
	GetWhitelists(logger lager.Logger, groups []string) ([]models.IngressWhitelist, error)
}

type RulesList struct {
	Marshaler marshal.Marshaler
	Logger    lager.Logger
	Store     store
}

func (h *RulesList) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	logger := h.Logger.Session("list-rules")
	logger.Info("start")
	defer logger.Info("done")

	all, err := h.Store.List(logger)
	if err != nil {
		logger.Error("store-list", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err := h.Marshaler.Marshal(all)
	if err != nil {
		logger.Error("marshal-failed", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.Header().Set("content-type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.Write(payload)
}

type RulesAdd struct {
	Unmarshaler marshal.Unmarshaler
	Logger      lager.Logger
	Store       store
}

func readRule(unmarshaler marshal.Unmarshaler, req *http.Request) (models.Rule, error) {
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return models.Rule{}, err
	}

	var rule models.Rule
	err = unmarshaler.Unmarshal(payload, &rule)
	if err != nil {
		return models.Rule{}, err
	}

	if err := rule.Validate(); err != nil {
		return models.Rule{}, err
	}
	return rule, nil
}

func (h *RulesAdd) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	logger := h.Logger.Session("add-rule")
	logger.Info("start")
	defer logger.Info("done")

	rule, err := readRule(h.Unmarshaler, req)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Info("adding", lager.Data{"rule": rule})

	err = h.Store.Add(logger, rule)
	if err != nil {
		logger.Error("store-add", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusCreated)
}

type RulesDelete struct {
	Unmarshaler marshal.Unmarshaler
	Logger      lager.Logger
	Store       store
}

func (h *RulesDelete) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	logger := h.Logger.Session("delete-rule")
	logger.Info("start")
	defer logger.Info("done")

	rule, err := readRule(h.Unmarshaler, req)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Info("deleting", lager.Data{"rule": rule})

	err = h.Store.Delete(logger, rule)
	if err != nil {
		logger.Error("store-delete", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp.WriteHeader(http.StatusNoContent)
}
