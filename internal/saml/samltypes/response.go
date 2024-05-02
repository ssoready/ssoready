package samltypes

import (
	"encoding/xml"
	"time"
)

type Response struct {
	XMLName   xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol Response"`
	Assertion struct {
		XMLName   xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
		Signature struct {
			XMLName    xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# Signature"`
			SignedInfo struct {
				XMLName                xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# SignedInfo"`
				CanonicalizationMethod struct {
					XMLName   xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# CanonicalizationMethod"`
					Algorithm string   `xml:"Algorithm,attr"`
				} `xml:"CanonicalizationMethod"`
				SignatureMethod struct {
					XMLName   xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# SignatureMethod"`
					Algorithm string   `xml:"Algorithm,attr"`
				} `xml:"SignatureMethod"`
				Reference struct {
					XMLName    xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# Reference"`
					Transforms struct {
						XMLName   xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# Transforms"`
						Transform []struct {
							XMLName             xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# Transform"`
							Algorithm           string   `xml:"Algorithm,attr"`
							InclusiveNamespaces struct {
								XMLName    xml.Name `xml:"http://www.w3.org/2001/10/xml-exc-c14n# InclusiveNamespaces"`
								PrefixList string   `xml:"PrefixList,attr"`
							} `xml:"InclusiveNamespaces"`
						} `xml:"Transform"`
					} `xml:"Transforms"`
					DigestMethod struct {
						XMLName   xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# DigestMethod"`
						Algorithm string   `xml:"Algorithm,attr"`
					} `xml:"DigestMethod"`
					DigestValue string
				} `xml:"Reference"`
			} `xml:"SignedInfo"`
			SignatureValue string
		} `xml:"Signature"`
		Issuer struct {
			XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
			Name    string   `xml:",chardata"`
		} `xml:"Issuer"`
		Subject struct {
			XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject"`
			NameID  struct {
				XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
				Value   string   `xml:",chardata"`
			} `xml:"NameID"`
			SubjectConfirmation struct {
				XMLName                 xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmation"`
				SubjectConfirmationData struct {
					XMLName      xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmationData"`
					InResponseTo string   `xml:"InResponseTo,attr"`
				} `xml:"SubjectConfirmationData"`
			} `xml:"SubjectConfirmation"`
		} `xml:"Subject"`
		Conditions struct {
			XMLName             xml.Name  `xml:"urn:oasis:names:tc:SAML:2.0:assertion Conditions"`
			NotBefore           time.Time `xml:"NotBefore,attr"`
			NotOnOrAfter        time.Time `xml:"NotOnOrAfter,attr"`
			AudienceRestriction struct {
				XMLName  xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion AudienceRestriction"`
				Audience struct {
					XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Audience"`
					Name    string   `xml:",chardata"`
				}
			}
		} `xml:"Conditions"`
		AttributeStatement struct {
			XMLName    xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeStatement"`
			Attributes []struct {
				XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Attribute"`
				Name    string   `xml:"Name,attr"`
				Value   string   `xml:"AttributeValue"`
			} `xml:"Attribute"`
		} `xml:"AttributeStatement"`
		AuthnStatement struct {
			XMLName      xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnStatement"`
			SessionIndex string   `xml:"SessionIndex,attr"`
		} `xml:"AuthnStatement"`
	} `xml:"Assertion"`
}
