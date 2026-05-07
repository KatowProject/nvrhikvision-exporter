package client

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/icholy/digest"
)

const (
	defaultTimeout    = 5 * time.Second
	defaultMaxRetries = 2
	retryDelay        = 500 * time.Millisecond
)

type Client struct {
	BaseURL    string
	HTTP       *http.Client
	MaxRetries int
}

func NewClient(ip, user, pass string) *Client {
	return &Client{
		BaseURL: fmt.Sprintf("http://%s", ip),
		HTTP: &http.Client{
			Transport: &digest.Transport{
				Username: user,
				Password: pass,
			},
			Timeout: defaultTimeout,
		},
		MaxRetries: defaultMaxRetries,
	}
}

func (c *Client) FetchXML(endpoint string, target interface{}) error {
	url := c.BaseURL + endpoint

	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
		}

		resp, err := c.HTTP.Get(url)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
		}

		return xml.NewDecoder(resp.Body).Decode(target)
	}

	return fmt.Errorf("after %d attempts: %w", c.MaxRetries+1, lastErr)
}
