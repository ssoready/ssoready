<p align="center">
<img src="https://github.com/ucarion/documentation/blob/main/Frame%2024%20(2).png?raw=true#gh-light-mode-only">
<img src="https://github.com/ucarion/documentation/blob/main/Frame%2025%20(2).png?raw=true#gh-dark-mode-only">
</p>

# SSOReady

We're building dev tools for implementing Enterprise SSO. You can use SSOReady to add SAML support to your product this
afternoon, for free, forever. You can think of us as an open source alternative to products like Auth0 or WorkOS.

* MIT-Licensed
* Self-hosted, or free at [app.ssoready.com](https://app.ssoready.com)
* Keeps you in control of your users database
* [Well-documented](https://ssoready.com/docs), straightforward implementation
* [Python](https://github.com/ssoready/ssoready-python) and
  [TypeScript/Node.js](https://github.com/ssoready/ssoready-typescript) SDKs, more in development

## Documentation

For full documentation, check out https://ssoready.com/docs.

At a super high level, all it takes to add SAML to your product is to:

1. Sign up on [app.ssoready.com](https://app.ssoready.com) for free
2. From your login page, call the `getRedirectUrl` endpoint when you want a user to sign in with SAML
3. Your user gets redirected back to a callback page you choose, e.g. `your-app.com/ssoready-callback?saml_access_code=...`. You
   call `redeemSamlAccessCode` with the `saml_access_code` and log them in.

Calling the `getRedirectUrl` endpoint looks like this in TypeScript:

```typescript
// this is how you implement a "Sign in with SSO" button
const { redirectUrl } = await ssoready.saml.getSamlRedirectUrl({
  // the ID of the organization/workspace/team (whatever you call it)
  // you want to log the user into
  organizationExternalId: "..."
});

// redirect the user to `redirectUrl`...
```

And `redeemSamlAccessCode` looks like this:

```typescript
// this goes in your handler for POST /ssoready-callback
const { email, organizationExternalId } = await ssoready.saml.redeemSamlAccessCode({
    samlAccessCode: "saml_access_code_..."
});

// log the user in as `email` inside `organizationExternalId`...
```

Check out [the quickstart](https://ssoready.com/docs) for the details spelled out more concretely. The whole point of
this project is to make enterprise SSO super obvious and easy.

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
