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

	if err := c.gqlClient().Query(ctx, &q, map[string]any{
		"appId":    req.AppID,
		"hostname": req.Hostname,
	}); err != nil {
		return nil, fmt.Errorf("graphql: %w", err)
	}

	return &GetCertificateResponse{Certificate: q.App.Certificate}, nil
}

type CheckCertificateRequest struct {
	AppID    string
	Hostname string
}

type CheckCertificateResponse struct {
	Certificate Certificate
}

func (c *Client) CheckCertificate(ctx context.Context, req *CheckCertificateRequest) (*CheckCertificateResponse, error) {
	var q struct {
		App struct {
			Certificate struct {
				Check      bool
				Configured bool
			} `graphql:"certificate(hostname: $hostname)"`
		} `graphql:"app(id: $appId)"`
	}

	if err := c.gqlClient().Query(ctx, &q, map[string]any{
		"appId":    req.AppID,
		"hostname": req.Hostname,
	}); err != nil {
		return nil, fmt.Errorf("graphql: %w", err)
	}

	return &CheckCertificateResponse{
		Certificate: Certificate{
			Configured: q.App.Certificate.Configured,
		},
	}, nil
}

type AddCertificateRequest struct {
	AppID    string
	Hostname string
}

type AddCertificateResponse struct {
	Certificate Certificate
}

func (c *Client) AddCertificate(ctx context.Context, req *AddCertificateRequest) (*AddCertificateResponse, error) {
	var m struct {
		AddCertificate struct {
			Certificate Certificate
		} `graphql:"addCertificate(appId: $appId, hostname: $hostname)"`
	}

	if err := c.gqlClient().Mutate(ctx, &m, map[string]any{
		"appId":    graphql.ID(req.AppID),
		"hostname": req.Hostname,
	}); err != nil {
		return nil, fmt.Errorf("graphql: %w", err)
	}

	return &AddCertificateResponse{Certificate: m.AddCertificate.Certificate}, nil
}

func (c *Client) gqlClient() *graphql.Client {
	httpClient := &http.Client{
		Transport: &authRoundTripper{
			BearerToken:  c.APIKey,
			RoundTripper: c.HTTPClient.Transport,
		},
		CheckRedirect: c.HTTPClient.CheckRedirect,
		Jar:           c.HTTPClient.Jar,
		Timeout:       c.HTTPClient.Timeout,
	}

	return graphql.NewClient("https://api.fly.io/graphql", httpClient)
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
