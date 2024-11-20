![](https://i.imgur.com/OhtkhbJ.png)

<div align="center">
  <h1>SSOReady</h1>
  <a href="https://github.com/ssoready/ssoready-typescript"><img src="https://img.shields.io/npm/v/ssoready.svg?style=flat&color=ECDC68" /></a>
  <a href="https://github.com/ssoready/ssoready-python"><img src="https://img.shields.io/pypi/v/ssoready.svg?style=flat" /></a>
  <a href="https://github.com/ssoready/ssoready-go"><img src="https://img.shields.io/github/v/tag/ssoready/ssoready-go?style=flat&label=golang&color=%23007D9C" /></a>
  <a href="https://github.com/ssoready/ssoready-csharp"><img src="https://img.shields.io/nuget/v/SSOReady.Client?style=flat&color=004880" /></a>
  <a href="https://github.com/ssoready/ssoready-ruby"><img src="https://img.shields.io/gem/v/ssoready?style=flat&color=EE3F2D" /></a>
  <a href="https://github.com/ssoready/ssoready-php"><img src="https://img.shields.io/packagist/v/ssoready/ssoready?style=flat&color=F28D1A" /></a>
  <a href="https://github.com/ssoready/ssoready/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" /></a>
  <a href="https://github.com/ssoready/ssoready/stargazers"><img src="https://img.shields.io/github/stars/ssoready/ssoready?style=flat&logo=github&color=white" /></a>
  <br />
  <br />
  <a href="https://ssoready.com/docs/saml/saml-quickstart">SAML Quickstart</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ssoready.com/docs/scim/scim-quickstart">SCIM Quickstart</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ssoready.com">Website</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ssoready.com/docs">Docs</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ssoready.com/blog">Blog</a>
  <br />
  <hr />
</div>

## What is SSOReady?

[SSOReady](https://ssoready.com) ([YC
W24](https://www.ycombinator.com/companies/ssoready)) is an **open-source,
straightforward** way to add SAML and SCIM support to your product:

* **[SSOReady SAML](https://ssoready.com/docs/saml/saml-quickstart)**: Everything you need to add SAML ("Enterprise SSO") to your product today.
* **[SSOReady SCIM](https://ssoready.com/docs/scim/scim-quickstart)**: Everything you need to add SCIM ("Enterprise Directory Sync") to your product today.
* **[Self-serve Setup UI](https://ssoready.com/docs/idp-configuration/enabling-self-service-configuration-for-your-customers)**:
  A hosted UI your customers use to onboard themselves onto SAML and/or
  SCIM.

**With SSOReady, you're in control:**

* SSOReady can be used in *any* application, regardless of what stack you use.
  We provide language-specific SDKs as thin wrappers over a [straightforward
  HTTP
  API](https://ssoready.com/docs/api-reference/saml/redeem-saml-access-code):
  * [SSOReady-TypeScript](https://github.com/ssoready/ssoready-typescript)
  * [SSOReady-Python](https://github.com/ssoready/ssoready-python)
  * [SSOReady-Go](https://github.com/ssoready/ssoready-go)
* SSOReady is just an authentication middleware layer. SSOReady doesn’t "own" your users or require any changes to your users database.
* You can use our cloud-hosted instance or [self-host yourself](https://ssoready.com/docs/self-hosting-ssoready), with the Enterprise plan giving you SLA'd support either way. 

**SSOReady can be extended with these products, available on the [Enterprise plan](https://ssoready.com/pricing):**

* [Custom Domains & Branding](https://ssoready.com/docs/ssoready-concepts/environments#custom-domains): Run
  SSOReady on a domain you control, and make your entire SAML/SCIM experience on-brand. 
* [Management API](https://ssoready.com/docs/management-api): Completely automate everything about SAML
  and SCIM programmatically at scale.
* [Enterprise Support](https://ssoready.com/pricing): SLA'd support, including for self-hosted deployments.

## Getting started

The fastest way to get started with SSOReady is to follow the quickstart for
what you want to add support for:

* [SAML Quickstart](https://ssoready.com/docs/saml/saml-quickstart)
* [SCIM Quickstart](https://ssoready.com/docs/scim/scim-quickstart)

Most folks implement SAML and SCIM in an afternoon. It only takes two lines of
code.

## How SSOReady works

This section provides a high-level overview of how SSOReady works, and how it's possible to implement SAML and SCIM in
just an afternoon. For a more thorough introduction, visit the [SAML
quickstart](https://ssoready.com/docs/saml/saml-quickstart) or the [SCIM
quickstart](https://ssoready.com/docs/scim/scim-quickstart).

### SAML in two lines of code

SAML (aka "Enterprise SSO") consists of two steps: an *initiation* step where you redirect your users to their corporate
identity provider, and a *handling* step where you log them in once you know who they are.

To initiate logins, you'll use SSOReady's [Get SAML Redirect
URL](https://ssoready.com/docs/api-reference/saml/get-saml-redirect-url) endpoint:

```typescript
// this is how you implement a "Sign in with SSO" button
const { redirectUrl } = await ssoready.saml.getSamlRedirectUrl({
  // the ID of the organization/workspace/team (whatever you call it)
  // you want to log the user into
  organizationExternalId: "..."
});

// redirect the user to `redirectUrl`...
```

You can use whatever your preferred ID is for organizations (you might call them "workspaces" or "teams") as your
`organizationExternalId`. You configure those IDs inside SSOReady, and SSOReady handles keeping track of that
organization's SAML and SCIM settings.

To handle logins, you'll use SSOReady's [Redeem SAML Access
Code](https://ssoready.com/docs/api-reference/saml/redeem-saml-access-code) endpoint:

```typescript
// this goes in your handler for POST /ssoready-callback
const { email, organizationExternalId } = await ssoready.saml.redeemSamlAccessCode({
  samlAccessCode: "saml_access_code_..."
});

// log the user in as `email` inside `organizationExternalId`...
```

You configure the URL for your `/ssoready-callback` endpoint in SSOReady.

### SCIM in one line of code

SCIM (aka "Enterprise directory sync") is basically a way for you to get a list of your customer's employees offline.

To get a customer's employees, you'll use SSOReady's [List SCIM
Users](https://ssoready.com/docs/api-reference/scim/list-scim-users) endpoint:

```typescript
const { scimUsers, nextPageToken } = await ssoready.scim.listScimUsers({
  organizationExternalId: "my_custom_external_id"
});

// create users from each scimUser
for (const { email, deleted, attributes } of scimUsers) {
  // ...
}
```

## Philosophy

We believe everyone that sells software to businesses should support enterprise
SSO. It's a huge security win for your customers.

The biggest problem with enterprise SSO is that it's way too confusing. Most
open-source SAML libraries are underdocumented messes. Every time I've tried to
implement SAML, I was constantly looking for someone to just tell me what in the
_world_ I was supposed to concretely do.

We believe that more people will implement enterprise SSO if you make it obvious
and secure by default. We are obsessed with giving every developer clarity and
security here.

Also, we believe randomly pumping up prices on security software like this is
totally unacceptable. MIT-licensing the software gives you insurance against us
ever doing that. Do whatever you want with the code. Fork us if we ever
misbehave.

## Security

If you have a security issue to report, please contact us at
security@ssoready.com.
