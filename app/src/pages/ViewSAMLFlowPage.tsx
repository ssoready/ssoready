import { useParams } from "react-router";
import { disableQuery, useQuery } from "@connectrpc/connect-query";
import {
  appGetSAMLConnection,
  appGetSAMLFlow,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import React from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import moment from "moment";
import formatXml from "xml-formatter";
import hljs from "highlight.js/lib/core";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Link } from "react-router-dom";
import { SAMLFlowStatus } from "@/gen/ssoready/v1/ssoready_pb";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { OctagonX } from "lucide-react";
import { Title } from "@/components/Title";
import { DocsLink } from "@/components/DocsLink";
import { InfoTooltip } from "@/components/InfoTooltip";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

export function ViewSAMLFlowPage() {
  const { environmentId, organizationId, samlConnectionId, samlFlowId } =
    useParams();
  const { data: samlFlow } = useQuery(appGetSAMLFlow, {
    id: samlFlowId,
  });
  const { data: samlConnection } = useQuery(appGetSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <div className="flex flex-col gap-8">
      <Title title="SAML Login Flow" />

      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link
                to={`/environments/${environmentId}/organizations/${organizationId}`}
              >
                {organizationId}
              </Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link
                to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}`}
              >
                {samlConnectionId}
              </Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />

          <BreadcrumbItem>
            <BreadcrumbPage>{samlFlowId}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>SAML Login Flow</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {samlFlowId}
                </span>
              </div>

              <CardDescription>
                A SAML login flow is a single time one of your users attempted
                to log in via SAML.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows" />
              </CardDescription>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-5 gap-y-2">
            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              Started
              <InfoTooltip>
                When the login process began.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows#start-time" />
              </InfoTooltip>
            </div>
            <div className="text-sm col-span-3">
              {samlFlow?.createTime &&
                moment(samlFlow.createTime.toDate()).format()}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              Last Activity
              <InfoTooltip>
                The last time any progress has happened on the SAML login flow.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows#last-activity-time" />
              </InfoTooltip>
            </div>
            <div className="text-sm col-span-3">
              {samlFlow?.updateTime &&
                moment(samlFlow.updateTime.toDate()).format()}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              Status
              <InfoTooltip>
                Whether the SAML login flow is in progress, succeeded, or
                failed.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows#status" />
              </InfoTooltip>
            </div>
            <div className="text-sm col-span-3">
              {samlFlow?.status ===
                SAMLFlowStatus.SAML_FLOW_STATUS_IN_PROGRESS && (
                <Badge variant="secondary">In progress</Badge>
              )}
              {samlFlow?.status === SAMLFlowStatus.SAML_FLOW_STATUS_FAILED && (
                <Badge variant="destructive">Failed</Badge>
              )}
              {samlFlow?.status ===
                SAMLFlowStatus.SAML_FLOW_STATUS_SUCCEEDED && (
                <Badge>Succeeded</Badge>
              )}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              State
              <InfoTooltip>
                The <code>state</code> you provided when getting a SAML redirect
                URL.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows#state" />
              </InfoTooltip>
            </div>
            <div className="col-span-3">
              {samlFlow?.state ? (
                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {samlFlow.state}
                </span>
              ) : (
                <span className="text-sm text-muted-foreground">None</span>
              )}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              User Email
              <InfoTooltip>
                Your end user's email address.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows#user-email" />
              </InfoTooltip>
            </div>
            <div className="text-sm col-span-3">
              {samlFlow?.email || (
                <span className="text-sm text-muted-foreground">None</span>
              )}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              User Attributes
              <InfoTooltip>
                Additional user attributes the IDP passed along.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows#user-attributes" />
              </InfoTooltip>
            </div>
            <div className="text-xs col-span-3">
              <span className="font-mono bg-gray-100 py-1 px-2 rounded-sm">
                {JSON.stringify(samlFlow?.attributes)}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {samlFlow?.error.case && (
        <Alert variant="destructive" className="bg-white shadow-sm">
          <OctagonX className="h-4 w-4" />
          <AlertTitle>This SAML flow was rejected by SSOReady</AlertTitle>

          {samlFlow.error.case === "samlConnectionNotConfigured" && (
            <AlertDescription>
              <p>
                You haven't finished configuring the SAML connection. Without an
                IDP Entity ID, IDP Redirect URL, and IDP Certificate, SSOReady
                cannot securely authenticate SAML logins.
              </p>

              <p className="mt-4">
                To fix this, you need to update{" "}
                <Link
                  className="underline underline-offset-4"
                  to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}`}
                >
                  the SAML connection
                </Link>{" "}
                to have an IDP Entity ID, IDP Redirect URL, and IDP Certificate.
              </p>
            </AlertDescription>
          )}

          {samlFlow.error.case ===
            "environmentOauthRedirectUriNotConfigured" && (
            <AlertDescription>
              <p>
                You haven't configured an OAuth Redirect URI on your
                environment, but this login flow used{" "}
                <a
                  className="underline underline-offset-4"
                  href="https://ssoready.com/docs/saml-over-oauth-saml-nextauth-integration"
                >
                  SAML-over-OAuth
                </a>
                .
              </p>

              <p className="mt-4">
                To fix this, you need to update the{" "}
                <Link
                  className="underline underline-offset-4"
                  to={`/environments/${environmentId}`}
                >
                  environment
                </Link>{" "}
                to have an OAuth Redirect URI configured.
              </p>
            </AlertDescription>
          )}

          {samlFlow.error.case === "unsignedAssertion" && (
            <AlertDescription>
              <p>
                Your customer's identity provider provided a SAML assertion that
                doesn't have a signature.
              </p>

              <p className="mt-4">
                Your customer's IT admin needs to enable assertion signing. They
                may have enable signing of the entire SAML response; SSOReady
                requires specifically the assertion be signed.
              </p>
            </AlertDescription>
          )}

          {samlFlow.error.case === "badIssuer" && (
            <AlertDescription>
              <p>
                You've configured the Identity Provider Entity ID as{" "}
                <span className="font-semibold">
                  {samlConnection?.idpEntityId}
                </span>
                , but your customer's identity provider provided{" "}
                <span className="font-semibold">{samlFlow?.error?.value}</span>.
              </p>

              <p className="mt-4">
                If you believe this login is legitimate, you need to update{" "}
                <Link
                  className="underline underline-offset-4"
                  to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}`}
                >
                  the SAML connection's
                </Link>{" "}
                Identity Provider Entity ID to{" "}
                <span className="font-semibold">{samlFlow.error.value}</span>.
                If you don't believe this login is legitimate, you don't need to
                do anything.
              </p>
            </AlertDescription>
          )}

          {samlFlow.error.case === "badAudience" && (
            <AlertDescription>
              <p>
                Your customer's identity provider provided a Service Provider
                Entity ID of{" "}
                <span className="font-semibold">{samlFlow.error.value}</span>,
                which is not the correct value.
              </p>

              <p className="mt-4">
                Your customer's IT admin needs to change the value to{" "}
                <span className="font-semibold">
                  {samlConnection?.spEntityId}
                </span>
                .
              </p>
            </AlertDescription>
          )}

          {samlFlow.error.case === "badSignatureAlgorithm" && (
            <AlertDescription>
              <p>
                Your customer's identity provider provided a SAML signature
                algorithm of{" "}
                <span className="font-semibold">{samlFlow.error.value}</span>,
                which SSOReady does not accept.
              </p>

              <p className="mt-4">
                Your customer's IT admin needs to change the value to{" "}
                <span className="font-semibold">
                  http://www.w3.org/2001/04/xmldsig-more#rsa-sha256
                </span>
                , often simply displayed as{" "}
                <span className="font-semibold">RSA-SHA256</span>.
              </p>
            </AlertDescription>
          )}

          {samlFlow.error.case === "badDigestAlgorithm" && (
            <AlertDescription>
              <p>
                Your customer's identity provider provided a SAML digest
                algorithm of{" "}
                <span className="font-semibold">{samlFlow.error.value}</span>,
                which SSOReady does not accept.
              </p>

              <p className="mt-4">
                Your customer's IT admin needs to change the value to{" "}
                <span className="font-semibold">
                  http://www.w3.org/2001/04/xmlenc#sha256
                </span>
                , often simply displayed as{" "}
                <span className="font-semibold">SHA256</span>.
              </p>
            </AlertDescription>
          )}

          {samlFlow.error.case === "badCertificate" && (
            <AlertDescription>
              <p>
                This request provided an incorrect SAML assertion signing
                certificate.
              </p>

              <Collapsible>
                <CollapsibleTrigger>
                  <p className="mt-4">
                    You've configured the the IDP Certificate to be: (click to
                    toggle)
                  </p>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <pre className="mt-2 inline-block py-2 px-4 rounded bg-red-100">
                    {samlConnection?.idpCertificate}
                  </pre>
                </CollapsibleContent>
              </Collapsible>

              <Collapsible>
                <CollapsibleTrigger>
                  <p className="mt-4">
                    This SAML request provided the following IDP Certificate:
                    (click to toggle)
                  </p>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <pre className="mt-2 inline-block py-2 px-4 rounded bg-red-100">
                    {samlFlow.error.value}
                  </pre>
                </CollapsibleContent>
              </Collapsible>

              <p className="mt-4">
                If you believe this login is legitimate, you need to update{" "}
                <Link
                  className="underline underline-offset-4"
                  to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}`}
                >
                  the SAML connection's
                </Link>{" "}
                Identity Provider Certificate to be the one provided on this
                request. If you don't believe this login is legitimate, you
                don't need to do anything.
              </p>
            </AlertDescription>
          )}

          {samlFlow.error.case === "emailOutsideOrganizationDomains" && (
            <AlertDescription>
              Your customer's identity provider returned a{" "}
              <span className="font-semibold">{samlFlow.error.value}</span>{" "}
              email, which is outside of the organization allowed domains.
            </AlertDescription>
          )}
        </Alert>
      )}

      <div className="relative">
        <span
          className="absolute left-5 top-5 -ml-px w-0.5 bg-gray-200 -z-10 h-[calc(100%-20px)]"
          aria-hidden="true"
        />
        <div className="flex flex-col gap-y-8">
          {samlFlow?.getRedirectTime && (
            <Card>
              <div className="p-6 space-y-2">
                <div className="flex gap-x-2">
                  <span className="text-sm text-muted-foreground">
                    {moment(samlFlow.getRedirectTime.toDate()).format()}
                  </span>
                  <span className="text-sm font-semibold">
                    Requested SAML Redirect URL
                    <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows#requested-saml-redirect-url" />
                  </span>
                </div>

                {samlFlow.authRedirectUrl && (
                  <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm max-w-full overflow-auto">
                    {samlFlow.authRedirectUrl}
                  </div>
                )}
              </div>
            </Card>
          )}
          {samlFlow?.initiateTime && (
            <Card>
              <div className="p-6 space-y-2">
                <div className="flex gap-x-2">
                  <span className="text-sm text-muted-foreground">
                    {moment(samlFlow.initiateTime.toDate()).format()}
                  </span>
                  <span className="text-sm font-semibold">
                    Initiated SAML Flow
                    <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows#initiated-saml-flow" />
                  </span>
                </div>

                <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm max-w-full overflow-auto">
                  <code>
                    <pre
                      dangerouslySetInnerHTML={{
                        __html: hljs.highlight(
                          formatXml(samlFlow.initiateRequest),
                          {
                            language: "xml",
                          },
                        ).value,
                      }}
                    />
                  </code>
                </div>
              </div>
            </Card>
          )}
          {samlFlow?.receiveAssertionTime && (
            <Card>
              <div className="p-6 space-y-2">
                <div className="flex gap-x-2">
                  <span className="text-sm text-muted-foreground">
                    {moment(samlFlow.receiveAssertionTime.toDate()).format()}
                  </span>
                  <span className="text-sm font-semibold">
                    Received SAML Assertion
                    <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows#received-saml-assertion" />
                  </span>
                </div>

                {samlFlow.assertion && (
                  <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm max-w-full overflow-auto">
                    <code>
                      <pre
                        dangerouslySetInnerHTML={{
                          __html: hljs.highlight(
                            formatXml(samlFlow.assertion),
                            {
                              language: "xml",
                            },
                          ).value,
                        }}
                      />
                    </code>
                  </div>
                )}
              </div>
            </Card>
          )}
          {samlFlow?.redeemTime && (
            <Card>
              <div className="p-6 space-y-2">
                <div className="flex gap-x-2">
                  <span className="text-sm text-muted-foreground">
                    {moment(samlFlow.redeemTime.toDate()).format()}
                  </span>
                  <span className="text-sm font-semibold">
                    Redeemed SAML Access Code
                    <DocsLink to="https://ssoready.com/docs/ssoready-concepts/saml-login-flows#redeemed-saml-access-code" />
                  </span>
                </div>

                <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm max-w-full overflow-auto">
                  <code>
                    <pre
                      dangerouslySetInnerHTML={{
                        __html: hljs.highlight(
                          JSON.stringify(
                            JSON.parse(samlFlow.redeemResponse),
                            null,
                            4,
                          ),
                          {
                            language: "json",
                          },
                        ).value,
                      }}
                    />
                  </code>
                </div>
              </div>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
