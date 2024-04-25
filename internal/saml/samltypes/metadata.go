package samltypes

import "encoding/xml"

type Metadata struct {
	XMLName          xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntityDescriptor"`
	EntityID         string   `xml:"entityID,attr"`
	IDPSSODescriptor struct {
		XMLName       xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:metadata IDPSSODescriptor"`
		KeyDescriptor struct {
			XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:metadata KeyDescriptor"`
			KeyInfo struct {
				XMLName  xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# KeyInfo"`
				X509Data struct {
					XMLName         xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# X509Data"`
					X509Certificate struct {
						XMLName xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# X509Certificate"`
						Value   string   `xml:",chardata"`
					} `xml:"X509Certificate"`
				} `xml:"X509Data"`
			} `xml:"KeyInfo"`
		} `xml:"KeyDescriptor"`
		SingleSignOnServices []struct {
			XMLName  xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:metadata SingleSignOnService"`
			Binding  string   `xml:"Binding,attr"`
			Location string   `xml:"Location,attr"`
		} `xml:"SingleSignOnService"`
	} `xml:"IDPSSODescriptor"`
}
