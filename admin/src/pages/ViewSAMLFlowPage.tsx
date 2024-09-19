import { useParams } from "react-router";
import { disableQuery, useQuery } from "@connectrpc/connect-query";
import {
  adminGetSAMLConnection,
  adminGetSAMLFlow,
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
import { LayoutMain } from "@/components/Layout";
import { Helmet } from "react-helmet";
import { useTitle } from "@/useTitle";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

export function ViewSAMLFlowPage() {
  const { samlConnectionId, samlFlowId } = useParams();
  const { data: samlFlow } = useQuery(adminGetSAMLFlow, {
    id: samlFlowId,
  });
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });
  const title = useTitle("SAML Flow");

  return (
    <LayoutMain>
      <Helmet>
        <title>{title}</title>
      </Helmet>

      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link to={`/saml/saml-connections/${samlConnectionId}`}>
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

      <Card className="mt-8">
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
                A SAML login flow is a single time someone attempted to log in
                via SAML.
              </CardDescription>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-5 gap-y-2">
            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              Started
            </div>
            <div className="text-sm col-span-3">
              {samlFlow?.samlFlow?.createTime &&
                moment(samlFlow.samlFlow.createTime.toDate()).format()}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              Last Activity
            </div>
            <div className="text-sm col-span-3">
              {samlFlow?.samlFlow?.updateTime &&
                moment(samlFlow.samlFlow.updateTime.toDate()).format()}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              Status
            </div>
            <div className="text-sm col-span-3">
              {samlFlow?.samlFlow?.status ===
                SAMLFlowStatus.SAML_FLOW_STATUS_IN_PROGRESS && (
                <Badge variant="secondary">In progress</Badge>
              )}
              {samlFlow?.samlFlow?.status ===
                SAMLFlowStatus.SAML_FLOW_STATUS_FAILED && (
                <Badge variant="destructive">Failed</Badge>
              )}
              {samlFlow?.samlFlow?.status ===
                SAMLFlowStatus.SAML_FLOW_STATUS_SUCCEEDED && (
                <Badge>Succeeded</Badge>
              )}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              State
            </div>
            <div className="col-span-3">
              {samlFlow?.samlFlow?.state ? (
                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {samlFlow.samlFlow.state}
                </span>
              ) : (
                <span className="text-sm text-muted-foreground">None</span>
              )}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              User Email
            </div>
            <div className="text-sm col-span-3">
              {samlFlow?.samlFlow?.email}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              User Attributes
            </div>
            <div className="text-xs col-span-3">
              <span className="font-mono bg-gray-100 py-1 px-2 rounded-sm">
                {JSON.stringify(samlFlow?.samlFlow?.attributes)}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {samlFlow?.samlFlow?.error.case && (
        <Alert variant="destructive" className="mt-8 bg-white shadow-sm">
          <OctagonX className="h-4 w-4" />
          <AlertTitle>This SAML flow was rejected</AlertTitle>

          {samlFlow.samlFlow.error.case === "unsignedAssertion" && (
            <AlertDescription>
              <p>
                Your identity provider provided a SAML assertion that doesn't
                have a signature.
              </p>

              <p className="mt-4">
                We require that SAML assertions be signed. This can be fixed by
                enabling SAML assertion signing in your identity provider
                settings.
              </p>
            </AlertDescription>
          )}

          {samlFlow.samlFlow.error.case === "badIssuer" && (
            <AlertDescription>
              <p>
                You've configured the Identity Provider Entity ID as{" "}
                <span className="font-semibold">
                  {samlConnection?.samlConnection?.idpEntityId}
                </span>
                , but your identity provider provided{" "}
                <span className="font-semibold">
                  {samlFlow.samlFlow?.error?.value}
                </span>
                .
              </p>

              <p className="mt-4">
                If you believe this login is legitimate, you need to{" "}
                <Link
                  className="underline underline-offset-4"
                  to={`/saml/saml-connections/${samlConnectionId}/setup`}
                >
                  set up the SAML connection's
                </Link>{" "}
                Identity Provider Entity ID to{" "}
                <span className="font-semibold">
                  {samlFlow.samlFlow.error.value}
                </span>
                .
              </p>
            </AlertDescription>
          )}

          {samlFlow.samlFlow.error.case === "badAudience" && (
            <AlertDescription>
              <p>
                Your identity provider provided a Service Provider Entity ID of{" "}
                <span className="font-semibold">
                  {samlFlow.samlFlow.error.value}
                </span>
                , which is not the correct value.
              </p>

              <p className="mt-4">
                You need to change the value to{" "}
                <span className="font-semibold">
                  {samlConnection?.samlConnection?.spEntityId}
                </span>{" "}
                in your identity provider.
              </p>
            </AlertDescription>
          )}

          {samlFlow.samlFlow.error.case === "badSignatureAlgorithm" && (
            <AlertDescription>
              <p>
                Your identity provider provided a SAML signature algorithm of{" "}
                <span className="font-semibold">
                  {samlFlow.samlFlow.error.value}
                </span>
                , which we do not accept.
              </p>

              <p className="mt-4">
                You need to change the value to{" "}
                <span className="font-semibold">
                  http://www.w3.org/2001/04/xmldsig-more#rsa-sha256
                </span>
                , often simply displayed as{" "}
                <span className="font-semibold">RSA-SHA256</span>.
              </p>
            </AlertDescription>
          )}

          {samlFlow.samlFlow.error.case === "badDigestAlgorithm" && (
            <AlertDescription>
              <p>
                Your identity provider provided a SAML digest algorithm of{" "}
                <span className="font-semibold">
                  {samlFlow.samlFlow.error.value}
                </span>
                , which we do not accept.
              </p>

              <p className="mt-4">
                You need to change the value to{" "}
                <span className="font-semibold">
                  http://www.w3.org/2001/04/xmlenc#sha256
                </span>
                , often simply displayed as{" "}
                <span className="font-semibold">SHA256</span>.
              </p>
            </AlertDescription>
          )}

          {samlFlow.samlFlow.error.case === "badCertificate" && (
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
                    {samlConnection?.samlConnection?.idpCertificate}
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
                    {samlFlow.samlFlow.error.value}
                  </pre>
                </CollapsibleContent>
              </Collapsible>

              <p className="mt-4">
                If you believe this login is legitimate, you need to{" "}
                <Link
                  className="underline underline-offset-4"
                  to={`/saml/saml-connections/${samlConnectionId}/setup`}
                >
                  set up the SAML connection's
                </Link>{" "}
                Identity Provider Certificate to be the one included on this
                login attempt.
              </p>
            </AlertDescription>
          )}

          {samlFlow.samlFlow.error.case ===
            "emailOutsideOrganizationDomains" && (
            <AlertDescription>
              Your identity provider returned a{" "}
              <span className="font-semibold">
                {samlFlow.samlFlow.error.value}
              </span>{" "}
              email, which is outside of your organization's allowed domains.
            </AlertDescription>
          )}
        </Alert>
      )}

      <div className="relative mt-8">
        <span
          className="absolute left-5 top-5 -ml-px w-0.5 bg-gray-200 -z-10 h-[calc(100%-20px)]"
          aria-hidden="true"
        />
        <div className="flex flex-col gap-y-8">
          {samlFlow?.samlFlow?.initiateTime && (
            <Card>
              <div className="p-6 space-y-2">
                <div className="flex gap-x-2">
                  <span className="text-sm text-muted-foreground">
                    {moment(samlFlow.samlFlow.initiateTime.toDate()).format()}
                  </span>
                  <span className="text-sm font-semibold">
                    Initiated SAML Flow
                  </span>
                </div>

                <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm max-w-full overflow-auto">
                  <code>
                    <pre
                      dangerouslySetInnerHTML={{
                        __html: hljs.highlight(
                          formatXml(samlFlow.samlFlow.initiateRequest),
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
          {samlFlow?.samlFlow?.receiveAssertionTime && (
            <Card>
              <div className="p-6 space-y-2">
                <div className="flex gap-x-2">
                  <span className="text-sm text-muted-foreground">
                    {moment(
                      samlFlow.samlFlow.receiveAssertionTime.toDate(),
                    ).format()}
                  </span>
                  <span className="text-sm font-semibold">
                    Received SAML Assertion
                  </span>
                </div>

                <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm max-w-full overflow-auto">
                  <code>
                    <pre
                      dangerouslySetInnerHTML={{
                        __html: hljs.highlight(
                          formatXml(samlFlow.samlFlow.assertion),
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
        </div>
      </div>
    </LayoutMain>
  );
}
