package flyio

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hasura/go-graphql-client"
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
	var q struct {
		App struct {
			Certificate Certificate `graphql:"certificate(hostname: $hostname)"`
		} `graphql:"app(id: $appId)"`
	}

	client := graphql.NewClient("https://api.fly.io/graphql", c.gqlClient())
	if err := client.Query(ctx, &q, map[string]any{
		"appId":    req.AppID,
		"hostname": req.Hostname,
	}); err != nil {
		return nil, fmt.Errorf("graphql: %w", err)
	}

	return &GetCertificateResponse{Certificate: q.App.Certificate}, nil
}

func (c *Client) gqlClient() *http.Client {
	return &http.Client{
		Transport: &authRoundTripper{
			BearerToken:  c.APIKey,
			RoundTripper: c.HTTPClient.Transport,
		},
		CheckRedirect: c.HTTPClient.CheckRedirect,
		Jar:           c.HTTPClient.Jar,
		Timeout:       c.HTTPClient.Timeout,
	}
}

type authRoundTripper struct {
	BearerToken  string
	RoundTripper http.RoundTripper
}

func (a *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.BearerToken))

	rt := a.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}
