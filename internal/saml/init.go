package saml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/url"
)

type InitRequest struct {
	RequestID      string
	IDPRedirectURL string
	SPEntityID     string
	RelayState     string
}

type InitResponse struct {
	URL             string
	InitiateRequest string
}

func Init(req *InitRequest) *InitResponse {
	redirectURL, err := url.Parse(req.IDPRedirectURL)
	if err != nil {
		panic(fmt.Errorf("parse idp redirect url: %w", err))
	}

	var samlReq samlRequest
	samlReq.ID = req.RequestID
	samlReq.Issuer.Name = req.SPEntityID
	samlReqData, err := xml.Marshal(samlReq)

	if err != nil {
		panic(fmt.Errorf("marshal AuthnRequest: %w", err))
	}

	query := redirectURL.Query()
	query.Set("SAMLRequest", base64.URLEncoding.EncodeToString(samlReqData))
	query.Set("RelayState", req.RelayState)
	redirectURL.RawQuery = query.Encode()

	return &InitResponse{URL: redirectURL.String(), InitiateRequest: string(samlReqData)}
}

type samlRequest struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol AuthnRequest"`
	ID      string   `xml:"ID,attr"`
	Issuer  struct {
		XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
		Name    string   `xml:",chardata"`
	} `xml:"Issuer"`
}
