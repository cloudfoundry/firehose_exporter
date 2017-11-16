package uaago

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	uaaUrl *url.URL
}

func NewClient(uaaUrl string) (*Client, error) {
	if len(uaaUrl) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	parsedURL, err := url.Parse(uaaUrl)
	if err != nil {
		return nil, err
	}

	return &Client{
		uaaUrl: parsedURL,
	}, nil
}

func (c *Client) GetAuthToken(username, password string, insecureSkipVerify bool) (string, error) {
	data := url.Values{
		"client_id":  {username},
		"grant_type": {"client_credentials"},
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth/token", c.uaaUrl), strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	request.SetBasicAuth(username, password)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	config := &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	tr := &http.Transport{TLSClientConfig: config}
	httpClient := &http.Client{Transport: tr}

	resp, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Received a status code %v", resp.Status)
	}

	jsonData := make(map[string]interface{})
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&jsonData)

	return fmt.Sprintf("%s %s", jsonData["token_type"], jsonData["access_token"]), err
}
