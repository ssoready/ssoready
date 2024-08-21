package flyio

import (
	"context"
	"net/http"
)

type Client struct {
	HTTPClient *http.Client
	APIKey     string
}

type GetCertificateRequest struct {
	AppID    string
	Hostname string
}

type GetCertificateResponse struct {
	Certificate Certificate
}

type Certificate struct {
	Configured bool
}

func (c *Client) GetCertificate(ctx context.Context, req *GetCertificateRequest) (*GetCertificateResponse, error) {
	client := 
}
