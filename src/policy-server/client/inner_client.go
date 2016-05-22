package client

import (
	"fmt"
	"net/http"
	"policy-server/models"

	"github.com/dghubble/sling"
)

func NewInnerClient(baseURL string, httpClient *http.Client) *InnerClient {
	slingClient := sling.New().Client(httpClient).Base(baseURL).Set("Accept", "application/json")
	return &InnerClient{
		slingClient: slingClient,
	}
}

type InnerClient struct {
	slingClient *sling.Sling
}

type filterQuery struct {
	Groups []string `url:"groups,comma"`
}

func (c *InnerClient) GetWhitelists(groupIDs []string) ([]models.IngressWhitelist, error) {
	var whitelists []models.IngressWhitelist

	resp, err := c.slingClient.New().
		Get("/whitelists").
		QueryStruct(filterQuery{
			Groups: groupIDs,
		}).
		Receive(&whitelists, nil)
	if err != nil {
		return nil, fmt.Errorf("list rules: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list rules: unexpected status code: %s", resp.Status)
	}

	return whitelists, nil
}
