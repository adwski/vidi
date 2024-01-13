package client

import (
	"context"
	"fmt"
	"strings"

	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	paramNameStatus   = "status"
	paramNameLocation = "location"
)

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
	return &Client{
		c:        resty.New(),
		logger:   cfg.Logger.With(zap.String("component", "video-api-client")),
		endpoint: strings.TrimSuffix(cfg.Endpoint, "/"),
	}
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
		Get(fmt.Sprintf("%s/service/search", c.endpoint))
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
	resp, err := req.Post(fmt.Sprintf("%s/serivce/%s/%s/%s", c.endpoint, videoID, param, value))
	return c.handleError(resp, response, err)
}

func (c *Client) UpdateVideo(videoID, status, location string) error {
	response, req := c.constructUpdateRequest()
	resp, err := req.
		SetBody(&model.VideoUpdateRequest{
			Status:   status,
			Location: location,
		}).
		Post(fmt.Sprintf("%s/serivce/%s", c.endpoint, videoID))
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
		return fmt.Errorf("video api responded with error status: %s", response.Error)
	}
	c.logger.Debug("video api request ok")
	return nil
}
