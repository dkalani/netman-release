package handlers

import (
	"lib/marshal"
	"net/http"
	"policy-server/models"
	"strings"

	"github.com/pivotal-golang/lager"
)

type Whitelists struct {
	Marshaler marshal.Marshaler
	Logger    lager.Logger
	Store     store
}

func (h *Whitelists) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	logger := h.Logger.Session("whitelists")
	logger.Info("start")
	defer logger.Info("done")

	queryValue := req.URL.Query().Get("groups")
	var groups []string
	if strings.TrimSpace(queryValue) != "" {
		groups = strings.Split(queryValue, ",")
	}
	all, err := h.Store.GetWhitelists(logger, groups)
	if err != nil {
		logger.Error("store-get-whitelists", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	if all == nil {
		all = []models.IngressWhitelist{}
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
