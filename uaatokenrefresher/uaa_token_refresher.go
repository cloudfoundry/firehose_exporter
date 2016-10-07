package uaatokenrefresher

import (
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/prometheus/common/log"
)

type UAATokenRefresher struct {
	url               string
	clientID          string
	clientSecret      string
	skipSSLValidation bool
	client            *uaago.Client
}

func New(
	url string,
	clientID string,
	clientSecret string,
	skipSSLValidation bool,
) (*UAATokenRefresher, error) {
	client, err := uaago.NewClient(url)
	if err != nil {
		return &UAATokenRefresher{}, err
	}

	return &UAATokenRefresher{
		url:               url,
		clientID:          clientID,
		clientSecret:      clientSecret,
		skipSSLValidation: skipSSLValidation,
		client:            client,
	}, nil
}

func (uaa *UAATokenRefresher) RefreshAuthToken() (string, error) {
	authToken, err := uaa.client.GetAuthToken(uaa.clientID, uaa.clientSecret, uaa.skipSSLValidation)
	if err != nil {
		log.Errorf("Error getting oauth token: %s. Please check your Client ID and Secret.", err.Error())
		return "", err
	}

	return authToken, nil
}
