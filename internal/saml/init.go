package saml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"time"
)

type InitRequest struct {
	RequestID  string
	SPEntityID string
	Now        time.Time
}

type InitResponse struct {
	SAMLRequest     string
	InitiateRequest string
}

func Init(req *InitRequest) *InitResponse {
	var samlReq samlRequest
	samlReq.ID = req.RequestID
	samlReq.Version = "2.0"
	samlReq.IssueInstant = req.Now.UTC().Truncate(time.Millisecond)
	samlReq.Issuer.Name = req.SPEntityID
	samlReqData, err := xml.Marshal(samlReq)

	if err != nil {
		panic(fmt.Errorf("marshal AuthnRequest: %w", err))
	}

	return &InitResponse{
		SAMLRequest:     base64.StdEncoding.EncodeToString(samlReqData),
		InitiateRequest: string(samlReqData),
	}
}

type samlRequest struct {
	XMLName      xml.Name  `xml:"urn:oasis:names:tc:SAML:2.0:protocol AuthnRequest"`
	ID           string    `xml:"ID,attr"`
	Version      string    `xml:"Version,attr"`
	IssueInstant time.Time `xml:"IssueInstant,attr"`
	Issuer       struct {
		XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
		Name    string   `xml:",chardata"`
	} `xml:"Issuer"`
}
