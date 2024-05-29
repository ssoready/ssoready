import { useParams } from "react-router";
import { disableQuery, useQuery } from "@connectrpc/connect-query";
import {
  getSAMLConnection,
  getSAMLFlow,
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

export function ViewSAMLFlowPage() {
  const { environmentId, organizationId, samlConnectionId, samlFlowId } =
    useParams();
  const { data: samlFlow } = useQuery(getSAMLFlow, {
    id: samlFlowId,
  });
  const { data: samlConnection } = useQuery(getSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <div className="flex flex-col gap-8">
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
                <CardTitle>SAML Flow</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {samlFlowId}
                </span>
              </div>

              <CardDescription>
                A SAML flow is a single time one of your users attempted to log
                in via SAML.
              </CardDescription>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-4 gap-y-2">
            <div className="text-sm col-span-1 text-muted-foreground">
              Started
            </div>
            <div className="text-sm col-span-3">
              {samlFlow?.createTime &&
                moment(samlFlow.createTime.toDate()).format()}
            </div>

            <div className="text-sm col-span-1 text-muted-foreground">
              Last Activity
            </div>
            <div className="text-sm col-span-3">
              {samlFlow?.updateTime &&
                moment(samlFlow.updateTime.toDate()).format()}
            </div>

            <div className="text-sm col-span-1 text-muted-foreground">
              Status
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

            <div className="text-sm col-span-1 text-muted-foreground">
              State
            </div>
            <div className="text-sm col-span-3">
              <span className="font-mono bg-gray-100 py-1 px-2 rounded-sm">
                {samlFlow?.state}
              </span>
            </div>

            <div className="text-sm col-span-1 text-muted-foreground">
              User Email
            </div>
            <div className="text-sm col-span-3">{samlFlow?.email}</div>

            <div className="text-sm col-span-1 text-muted-foreground">
              User Attributes
            </div>
            <div className="text-sm col-span-3">
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
                  </span>
                </div>

                <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm max-w-full overflow-auto">
                  {samlFlow.authRedirectUrl}
                </div>
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
                  </span>
                </div>

                <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm max-w-full overflow-auto">
                  <code>
                    <pre
                      dangerouslySetInnerHTML={{
                        __html: hljs.highlight(formatXml(samlFlow.assertion), {
                          language: "xml",
                        }).value,
                      }}
                    />
                  </code>
                </div>
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
