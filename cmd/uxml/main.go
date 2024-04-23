package main

import (
	"encoding/json"
	"fmt"

	"github.com/ssoready/ssoready/internal/uxml"
)

func main() {
	//s := `<foo bar="baz">qux<a>aaa<b>bbb</b></a></foo>`
	//s := `<foo bar="baz">text1<a b="c">text2</a>text3</foo>`
	s := `<?xml version="1.0" encoding="UTF-8"?>
<saml2p:Response Destination="http://localhost:8080/123" ID="id15382897449827171224409323"
                 IssueInstant="2023-11-16T21:20:00.399Z" Version="2.0"
                 xmlns:saml2p="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:xs="http://www.w3.org/2001/XMLSchema">
    <saml2:Issuer Format="urn:oasis:names:tc:SAML:2.0:nameid-format:entity"
                  xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">http://www.okta.com/exk9e5bqltmtkaQEy697
    </saml2:Issuer>
    <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
        <ds:SignedInfo>
            <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
            <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
            <ds:Reference URI="#id15382897449827171224409323">
                <ds:Transforms>
                    <ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
                    <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#">
                        <ec:InclusiveNamespaces PrefixList="xs" xmlns:ec="http://www.w3.org/2001/10/xml-exc-c14n#"/>
                    </ds:Transform>
                </ds:Transforms>
                <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
                <ds:DigestValue>AcePgS8ChmBavb2AyftTiU7yEkPInig8nt5kyMqn5q0=</ds:DigestValue>
            </ds:Reference>
        </ds:SignedInfo>
        <ds:SignatureValue>
            iq3qu5FkZ76SglwKZ2yRbBbfJqEwtBeLpNiyyQ1elVjTd9uJvm16V3EjYG5DnpNxf+Ao16eZn2CuamTLjYZJa1icN8VldF+d4oYo7qu+2aX3Z6GiCDGme7BpXuC0qObGJW86H3o5ZxOc8xuS0gOJaJg/JPP4COpDNMCfn/3e0Q1GGs5pTq9llbZJJwit8lI+G5SNO0CnUXQs47nlfVI5/qI7zzd7/1eFSp/AzVf/BOceQ1PjYH59A1c9dhzkwFUqINvryyNRX+2oGiKXfVCzIZglfkRqc4xfcZwcSZF4Arj5MuZFFLjf5OcaXkFci1C9jPpfvFC/YLDA4yx8Y1N1KA==
        </ds:SignatureValue>
        <ds:KeyInfo>
            <ds:X509Data>
                <ds:X509Certificate>MIIDqjCCApKgAwIBAgIGAYvaAJgxMA0GCSqGSIb3DQEBCwUAMIGVMQswCQYDVQQGEwJVUzETMBEG
                    A1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwNU2FuIEZyYW5jaXNjbzENMAsGA1UECgwET2t0YTEU
                    MBIGA1UECwwLU1NPUHJvdmlkZXIxFjAUBgNVBAMMDXRyaWFsLTU5NTQ5MzgxHDAaBgkqhkiG9w0B
                    CQEWDWluZm9Ab2t0YS5jb20wHhcNMjMxMTE2MjExODEzWhcNMzMxMTE2MjExOTEzWjCBlTELMAkG
                    A1UEBhMCVVMxEzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDVNhbiBGcmFuY2lzY28xDTAL
                    BgNVBAoMBE9rdGExFDASBgNVBAsMC1NTT1Byb3ZpZGVyMRYwFAYDVQQDDA10cmlhbC01OTU0OTM4
                    MRwwGgYJKoZIhvcNAQkBFg1pbmZvQG9rdGEuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
                    CgKCAQEAuRu7V88DgNWYMxCRXv27bpRg5JGiWkq9Bpnv+mrmNY8jq//hwYGczsftcjeXk8g6IWR3
                    mcJzmFFNU4K+9cn94bnpn7F8ASGWXtoPiokkftDCHqgX8GqeN9c1OTxu/FdT8vY40w08Jl7wmoCZ
                    mwtjcMn5tk4HFP901ycgJQWGC2YSYv8HYrqF9GD159FdEgRPHfMxczr3S2sGxAdUh2JbOck3cGAh
                    ODJHuxhaKZUb3zkqo9ZGCzIeQskOOjJ1twpj3SnJaeze7AFfB1wFS14bt6OajZHaikQR0nHGQ4xr
                    d/9DqR1wCFqqLWWfDD5wNWjUTFuq0FWNKWhIh3ti6lzt8QIDAQABMA0GCSqGSIb3DQEBCwUAA4IB
                    AQA/EN8XxNkYRz08He4gXs2dP6wXxxDhaq+w6V+4jDW+3w6SngpNH7/oK5Kjgtp57R4Q39QqwzPA
                    pOgKoyt9D6n5I1MI8wLL+DwhjPbF8kyj6jj4p73vEOBwLl7Ihw3mn4WVy0c6+jkEI3PD1efnk8ji
                    u27MnOx2M8zXh/FFE1j3AaPEjU5z99Yy3MnVM+NAn7nxJe+0/Px4RC3VChNULXfMLXoxrbsjT568
                    Wb2u4RCz/SFpM7Jey7XmjVzZEdHt9kmcnWJKfTg31HV4nG3uKG7FMqAAhsapZn8gnXziYeLL0+ja
                    VyhY+htTzahr+JR9yPKXQ2yzXXcHDR1JeE4DzwgG
                </ds:X509Certificate>
            </ds:X509Data>
        </ds:KeyInfo>
    </ds:Signature>
    <saml2p:Status xmlns:saml2p="urn:oasis:names:tc:SAML:2.0:protocol">
        <saml2p:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/>
    </saml2p:Status>
    <saml2:Assertion ID="id153828974522020275822991" IssueInstant="2023-11-16T21:20:00.399Z" Version="2.0"
                     xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion" xmlns:xs="http://www.w3.org/2001/XMLSchema">
        <saml2:Issuer Format="urn:oasis:names:tc:SAML:2.0:nameid-format:entity"
                      xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">http://www.okta.com/exk9e5bqltmtkaQEy697
        </saml2:Issuer>
        <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
            <ds:SignedInfo>
                <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
                <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
                <ds:Reference URI="#id153828974522020275822991">
                    <ds:Transforms>
                        <ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
                        <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#">
                            <ec:InclusiveNamespaces PrefixList="xs" xmlns:ec="http://www.w3.org/2001/10/xml-exc-c14n#"/>
                        </ds:Transform>
                    </ds:Transforms>
                    <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
                    <ds:DigestValue>TKAVTIn3t8z55tjMmUz9LNAqdr9GrX9jDs8MRh5XeyM=</ds:DigestValue>
                </ds:Reference>
            </ds:SignedInfo>
            <ds:SignatureValue>
                RDMyD66YnxkMZCLQLGMoW/U9ucIbEX1cyB7HhMKE0M8/WRbLtE4IRjCSX7XBFny58qnvi0IVSusVgHdNdt+yN4mEMin59Pr4uEygNQLYjd5pH44KbpuWsFaoUq6WvofiUMHNFcTbqxo8OlGG0dvoKwmc68VSosQjrC5sWQp4T6OO9aj/VjLkYWB6kOTmbiVmPutU5N5ligTq+fphfIkIuJrh0UzsYa7cRyRlIbCcGTXVHpTG6uVEqw9QRcinayQ2umqpmmMhSEx4PsnpUMxTF83DQhvRfLXY5cNNJ7mGu+liCQbiNmaUUCmpCjQPWZRasQ4LJlrtZoXbxLZLrw5suA==
            </ds:SignatureValue>
            <ds:KeyInfo>
                <ds:X509Data>
                    <ds:X509Certificate>MIIDqjCCApKgAwIBAgIGAYvaAJgxMA0GCSqGSIb3DQEBCwUAMIGVMQswCQYDVQQGEwJVUzETMBEG
                        A1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwNU2FuIEZyYW5jaXNjbzENMAsGA1UECgwET2t0YTEU
                        MBIGA1UECwwLU1NPUHJvdmlkZXIxFjAUBgNVBAMMDXRyaWFsLTU5NTQ5MzgxHDAaBgkqhkiG9w0B
                        CQEWDWluZm9Ab2t0YS5jb20wHhcNMjMxMTE2MjExODEzWhcNMzMxMTE2MjExOTEzWjCBlTELMAkG
                        A1UEBhMCVVMxEzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDVNhbiBGcmFuY2lzY28xDTAL
                        BgNVBAoMBE9rdGExFDASBgNVBAsMC1NTT1Byb3ZpZGVyMRYwFAYDVQQDDA10cmlhbC01OTU0OTM4
                        MRwwGgYJKoZIhvcNAQkBFg1pbmZvQG9rdGEuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
                        CgKCAQEAuRu7V88DgNWYMxCRXv27bpRg5JGiWkq9Bpnv+mrmNY8jq//hwYGczsftcjeXk8g6IWR3
                        mcJzmFFNU4K+9cn94bnpn7F8ASGWXtoPiokkftDCHqgX8GqeN9c1OTxu/FdT8vY40w08Jl7wmoCZ
                        mwtjcMn5tk4HFP901ycgJQWGC2YSYv8HYrqF9GD159FdEgRPHfMxczr3S2sGxAdUh2JbOck3cGAh
                        ODJHuxhaKZUb3zkqo9ZGCzIeQskOOjJ1twpj3SnJaeze7AFfB1wFS14bt6OajZHaikQR0nHGQ4xr
                        d/9DqR1wCFqqLWWfDD5wNWjUTFuq0FWNKWhIh3ti6lzt8QIDAQABMA0GCSqGSIb3DQEBCwUAA4IB
                        AQA/EN8XxNkYRz08He4gXs2dP6wXxxDhaq+w6V+4jDW+3w6SngpNH7/oK5Kjgtp57R4Q39QqwzPA
                        pOgKoyt9D6n5I1MI8wLL+DwhjPbF8kyj6jj4p73vEOBwLl7Ihw3mn4WVy0c6+jkEI3PD1efnk8ji
                        u27MnOx2M8zXh/FFE1j3AaPEjU5z99Yy3MnVM+NAn7nxJe+0/Px4RC3VChNULXfMLXoxrbsjT568
                        Wb2u4RCz/SFpM7Jey7XmjVzZEdHt9kmcnWJKfTg31HV4nG3uKG7FMqAAhsapZn8gnXziYeLL0+ja
                        VyhY+htTzahr+JR9yPKXQ2yzXXcHDR1JeE4DzwgG
                    </ds:X509Certificate>
                </ds:X509Data>
            </ds:KeyInfo>
        </ds:Signature>
        <saml2:Subject xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
            <saml2:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified">
                ulysse.carion@codomaindata.com
            </saml2:NameID>
            <saml2:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer">
                <saml2:SubjectConfirmationData NotOnOrAfter="2023-11-16T21:25:00.399Z"
                                               Recipient="http://localhost:8080/123"/>
            </saml2:SubjectConfirmation>
        </saml2:Subject>
        <saml2:Conditions NotBefore="2023-11-16T21:15:00.399Z" NotOnOrAfter="2023-11-16T21:25:00.399Z"
                          xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
            <saml2:AudienceRestriction>
                <saml2:Audience>http://localhost:8080/123</saml2:Audience>
            </saml2:AudienceRestriction>
        </saml2:Conditions>
        <saml2:AuthnStatement AuthnInstant="2023-11-16T21:20:00.026Z" SessionIndex="id1700169587132.854755065"
                              xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
            <saml2:AuthnContext>
                <saml2:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport
                </saml2:AuthnContextClassRef>
            </saml2:AuthnContext>
        </saml2:AuthnStatement>
        <saml2:AttributeStatement xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
            <saml2:Attribute Name="email" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified">
                <saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema"
                                      xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">
                    ulysse.carion@codomaindata.com
                </saml2:AttributeValue>
            </saml2:Attribute>
            <saml2:Attribute Name="firstName" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified">
                <saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema"
                                      xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Ulysse
                </saml2:AttributeValue>
            </saml2:Attribute>
            <saml2:Attribute Name="lastName" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified">
                <saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema"
                                      xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Carion
                </saml2:AttributeValue>
            </saml2:Attribute>
        </saml2:AttributeStatement>
    </saml2:Assertion>
</saml2p:Response>
`

	doc, err := uxml.Parse(s)
	if err != nil {
		panic(err)
	}
	d, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(d))

	//l, err := uxml.Lex.LexString("", s)
	//if err != nil {
	//	panic(err)
	//}
	//for {
	//	t, err := l.Next()
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println(nameof(t.Type), strconv.Quote(t.String()))
	//	if t.EOF() {
	//		break
	//	}
	//}
	//
	//doc, err := uxml.parser.ParseString("", s)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("%#v\n", doc)
}

//
//func nameof(t lexer.TokenType) string {
//	for k, v := range uxml.lex.Symbols() {
//		if v == t {
//			return k
//		}
//	}
//	return ""
//}
