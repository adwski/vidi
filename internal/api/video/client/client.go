package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"

	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	paramNameStatus   = "status"
	paramNameLocation = "location"
)

// Client is a Video API service-side client.
// It is used as a component in media services.
type Client struct {
	c        *resty.Client
	logger   *zap.Logger
	token    string
	endpoint string
}

type Config struct {
	Logger   *zap.Logger
	Endpoint string
	Token    string
}

func New(cfg *Config) *Client {
	c := &Client{
		c:        resty.New(),
		token:    cfg.Token,
		logger:   cfg.Logger.With(zap.String("component", "video-api-client")),
		endpoint: strings.TrimSuffix(cfg.Endpoint, "/"),
	}
	if len(c.token) == 0 {
		c.logger.Warn("starting with empty token")
	}
	return c
}

func (c *Client) GetUploadedVideos(ctx context.Context) ([]*model.Video, error) {
	var (
		errResponse    common.Response
		videosResponse = make([]*model.Video, 0)
	)
	resp, err := c.c.R().SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetAuthToken(c.token).
		SetError(&errResponse).
		SetResult(&videosResponse).
		SetBody(&model.ListRequest{
			Status: model.StatusUploaded,
		}).
		Post(fmt.Sprintf("%s/service/search", c.endpoint))
	if err != nil {
		return nil, c.handleError(resp, &errResponse, err)
	}
	return videosResponse, nil
}

func (c *Client) UpdateVideoStatus(videoID, param string) error {
	return c.makeUpdateRequest(videoID, paramNameStatus, param)
}

func (c *Client) UpdateVideoLocation(videoID, param string) error {
	return c.makeUpdateRequest(videoID, paramNameLocation, param)
}

func (c *Client) makeUpdateRequest(videoID, param, value string) error {
	response, req := c.constructUpdateRequest()
	resp, err := req.Put(fmt.Sprintf("%s/service/%s/%s/%s", c.endpoint, videoID, param, value))
	return c.handleError(resp, response, err)
}

func (c *Client) UpdateVideo(videoID, status, location string) error {
	response, req := c.constructUpdateRequest()
	spew.Dump(req.Token)
	resp, err := req.
		SetBody(&model.UpdateRequest{
			Status:   status,
			Location: location,
		}).
		Put(fmt.Sprintf("%s/service/%s", c.endpoint, videoID))
	return c.handleError(resp, response, err)
}

func (c *Client) constructUpdateRequest() (*common.Response, *resty.Request) {
	var response common.Response
	return &response, c.c.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(c.token).
		SetError(&response).
		SetResult(&response)
}

func (c *Client) handleError(resp *resty.Response, response *common.Response, err error) error {
	if err != nil {
		return fmt.Errorf("error while making video api request: %w", err)
	}
	if resp.IsError() {
		c.logger.Error("api call returned error",
			zap.String("status", resp.Status()),
			zap.String("api-error", response.Error))
		return fmt.Errorf("video api responded with error %s status", resp.Status())
	}
	c.logger.Debug("video api request ok")
	return nil
}
