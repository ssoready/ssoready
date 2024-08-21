package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	c := FlyClient{
		HTTPClient: http.DefaultClient,
		APIKey:     "FlyV1 fm2_lJPECAAAAAAAAzrNxBAT1jr27LyVYpvk9Fewrt4lwrVodHRwczovL2FwaS5mbHkuaW8vdjGUAJLOAAxEsx8Lk7lodHRwczovL2FwaS5mbHkuaW8vYWFhL3YxxDwZiiggBQVKQxk4tyN2e3ds4zlgFrazJi2PBEHns9Lw7iJ5z2k97P65i3ks9P7ZTsbJ/HBTY0bhSrdaFPbETobKzA45z59SR/1BEieQ7LQmMNYgqqATxPuTjHBRGqauPWxvRNP6hJS66WplbuGoGujIlDZkX0hEGCkI0G/aGbtQMjd1aYLfLn8by7c9o8Qg5N+yKMfuuQ0EGL0ZdrDksJ/QRJEf+2FBduqrD7YYJrw=,fm2_lJPETobKzA45z59SR/1BEieQ7LQmMNYgqqATxPuTjHBRGqauPWxvRNP6hJS66WplbuGoGujIlDZkX0hEGCkI0G/aGbtQMjd1aYLfLn8by7c9o8QQ6IxvxPaUCKvim4gF+/ILN8O5aHR0cHM6Ly9hcGkuZmx5LmlvL2FhYS92MZgEks5mxQO+zwAAAAEivSHcF84AC0zjCpHOAAtM4wzEECkFE5nZl3Jb/5jnmpyFJ5HEIE5yCjhZ0c4TQ6IeLopa9Fg/zJDSrFIINRM859tbQHGf",
	}

	//fmt.Println(c.GetAppCertificates(context.Background(), "authproxy-twilight-violet-4061"))
	//fmt.Println(c.AddCertificate(context.Background(), "authproxy-twilight-violet-4061", "auth.acmecorp.com"))
	data, err := c.GetAppCertificates(context.Background(), "authproxy-twilight-violet-4061", "auth.acmecorp.com")
	if err != nil {
		panic(err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(data); err != nil {
		panic(err)
	}
}

type FlyClient struct {
	HTTPClient *http.Client
	APIKey     string
}

func (f *FlyClient) GetAppCertificates(ctx context.Context, appID, hostname string) (any, error) {
	query := `query($appId: String!, $hostname: String!) {
  app(id: $appId) {
    certificate(hostname: $hostname) {
      configured
      acmeDnsConfigured
      acmeAlpnConfigured
      certificateAuthority
      createdAt
      dnsProvider
      dnsValidationInstructions
      dnsValidationHostname
      dnsValidationTarget
      hostname
      id
      source
      clientStatus
      issued {
        nodes {
          type
          expiresAt
        }
      }
    }
  }
}
`
	variables := map[string]interface{}{
		"appId":    appID,
		"hostname": hostname,
	}

	// put variables into req
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(map[string]interface{}{
		"query":     query,
		"variables": variables,
	}); err != nil {
		return nil, err
	}

	// run the query
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.fly.io/graphql", &body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.APIKey))

	res, err := f.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad status code: %v, body: %s", res.StatusCode, string(body))
	}

	// parse the response
	var data struct {
		Data struct {
			App struct {
				Certificate struct {
					Configured                bool
					ACMEDNSConfigured         bool
					ACMEALPNConfigured        bool
					CertificateAuthority      string
					CreatedAt                 string
					DNSProvider               string
					DNSValidationInstructions string
					DNSValidationHostname     string
					DNSValidationTarget       string
					Hostname                  string
					ID                        string
					Source                    string
					ClientStatus              string
					Issued                    struct {
						Nodes []struct {
							Type      string
							ExpiresAt string
						}
					}
				}
			}
		}
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (f *FlyClient) AddCertificate(ctx context.Context, appID, hostname string) (any, error) {
	mutation := `mutation($appId: ID!, $hostname: String!) {
    addCertificate(appId: $appId, hostname: $hostname) {
        certificate {
            configured
            acmeDnsConfigured
            acmeAlpnConfigured
            certificateAuthority
            certificateRequestedAt
            dnsProvider
            dnsValidationInstructions
            dnsValidationHostname
            dnsValidationTarget
            hostname
            id
            source
        }
    }
}`

	fmt.Println("appid", appID)
	variables := map[string]interface{}{
		"appId":    appID,
		"hostname": hostname,
	}

	// put variables into req
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(map[string]interface{}{
		"query":     mutation,
		"variables": variables,
	}); err != nil {
		return nil, err
	}

	// run the query
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.fly.io/graphql", &body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.APIKey))

	res, err := f.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	//if res.StatusCode != 200 {
	{
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad status code: %v, body: %s", res.StatusCode, string(body))
	}

	// parse the response
	var data struct {
		AddCertificate struct {
			Certificate struct {
				Configured                bool
				ACMEDNSConfigured         bool
				ACMEALPNConfigured        bool
				CertificateAuthority      string
				CreatedAt                 string
				DNSProvider               string
				DNSValidationInstructions string
				DNSValidationHostname     string
				DNSValidationTarget       string
				Hostname                  string
				ID                        string
				Source                    string
				ClientStatus              string
				Issued                    struct {
					Nodes []struct {
						Type      string
						ExpiresAt string
					}
				}
			}
		}
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	fmt.Println(data)
	return data, nil
}
