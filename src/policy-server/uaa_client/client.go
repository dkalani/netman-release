package uaa_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type BadUaaResponse struct {
	StatusCode      int
	UaaResponseBody string
}

func (r BadUaaResponse) Error() string {
	return fmt.Sprintf("bad uaa response: %d: %s", r.StatusCode, r.UaaResponseBody)
}

type Client struct {
	Host       string
	Name       string
	Secret     string
	HTTPClient httpClient
}

//go:generate counterfeiter -o ../fakes/http_client.go --fake-name HTTPClient . httpClient
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type CheckTokenResponse struct {
	Scope    []string `json:"scope"`
	UserName string   `json:"user_name"`
}

func (c *Client) GetName(token string) (string, error) {
	reqURL := fmt.Sprintf("%s/check_token?token=%s", c.Host, token)
	request, err := http.NewRequest("GET", reqURL, nil)
	request.SetBasicAuth(c.Name, c.Secret)

	resp, err := c.HTTPClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("http client: %s", err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %s", err)
	}

	if resp.StatusCode != 200 {
		err = BadUaaResponse{
			StatusCode:      resp.StatusCode,
			UaaResponseBody: string(respBytes),
		}
		return "", err
	}

	responseStruct := CheckTokenResponse{}
	err = json.Unmarshal(respBytes, &responseStruct)
	if err != nil {
		return "", fmt.Errorf("unmarshal json: %s", err)
	}
	if !hasRequiredScope(responseStruct.Scope, "network.admin") {
		return "", errors.New("network.admin scope not found")
	}

	return responseStruct.UserName, nil
}

func hasRequiredScope(scopes []string, requiredScope string) bool {
	for _, scope := range scopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}
