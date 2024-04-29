package client

import (
	"errors"
	"fmt"
	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/go-resty/resty/v2"
	"strings"
)

const jwtCookieName = "vidiSessID"

// Client is a User API client.
type Client struct {
	c        *resty.Client
	endpoint string
}

type Config struct {
	Endpoint string
}

func New(cfg *Config) *Client {
	return &Client{
		c:        resty.New(),
		endpoint: strings.TrimSuffix(cfg.Endpoint, "/"),
	}
}

func (c *Client) Login(username, password string) (string, error) {
	response, req := c.constructRequest()
	resp, err := req.SetBody(&model.UserRequest{
		Username: username,
		Password: password,
	}).Post(fmt.Sprintf("%s/login", c.endpoint))
	if err = c.handleError(resp, response, err); err != nil {
		return "", err
	}
	return getTokenFromCookies(resp)
}

func (c *Client) Register(username, password string) (string, error) {
	response, req := c.constructRequest()
	resp, err := req.SetBody(&model.UserRequest{
		Username: username,
		Password: password,
	}).Post(fmt.Sprintf("%s/register", c.endpoint))
	if err = c.handleError(resp, response, err); err != nil {
		return "", err
	}
	return getTokenFromCookies(resp)
}

func getTokenFromCookies(resp *resty.Response) (string, error) {
	for _, c := range resp.Cookies() {
		if c.Name == jwtCookieName {
			return c.Value, nil
		}
	}
	return "", errors.New("no jwt cookie found")
}

func (c *Client) constructRequest() (*common.Response, *resty.Request) {
	var response common.Response
	return &response, c.c.R().
		SetHeader("Content-Type", "application/json").
		SetError(&response).
		SetResult(&response)
}

func (c *Client) handleError(resp *resty.Response, response *common.Response, err error) error {
	if err != nil {
		return fmt.Errorf("error while making request: %w", err)
	}
	if resp.IsError() || response.Error != "" {
		return fmt.Errorf("user api responded with status: %s, error: '%s'", resp.Status(), response.Error)
	}
	return nil
}
