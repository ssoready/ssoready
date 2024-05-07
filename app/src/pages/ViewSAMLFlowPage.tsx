import { useMatch, useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import {
  getSAMLConnection,
  getSAMLFlow,
  listSAMLFlows,
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

export function ViewSAMLFlowPage() {
  const { samlFlowId } = useParams();
  const { data: samlFlow } = useQuery(getSAMLFlow, {
    id: samlFlowId,
  });

  return (
    <div className="flex flex-col gap-8">
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
              State
            </div>
            <div className="text-sm col-span-3">
              <span className="font-mono bg-gray-100 py-1 px-2 rounded-sm">
                {samlFlow?.state}
              </span>
            </div>

            <div className="text-sm col-span-1 text-muted-foreground">
              User ID
            </div>
            <div className="text-sm col-span-3">{samlFlow?.subjectIdpId}</div>

            <div className="text-sm col-span-1 text-muted-foreground">
              User Attributes
            </div>
            <div className="text-sm col-span-3">
              <span className="font-mono bg-gray-100 py-1 px-2 rounded-sm">
                {JSON.stringify(samlFlow?.subjectIdpAttributes)}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="relative">
        <span
          className="absolute left-5 top-5 -ml-px h-full w-0.5 bg-gray-200 -z-10"
          aria-hidden="true"
        />
        <div className="flex flex-col gap-y-8">
          {samlFlow?.getRedirectTime && (
            <Card>
              <div className="p-6 space-y-2">
                <div className="flex gap-x-2">
                  <span className="text-sm font-semibold">
                    {moment(samlFlow.getRedirectTime.toDate()).format()}
                  </span>
                  <span className="text-sm">Requested SAML Redirect URL</span>
                </div>

                <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {samlFlow.authRedirectUrl}
                </div>
              </div>
            </Card>
          )}
          {samlFlow?.initiateTime && (
            <Card>
              <div className="p-6 space-y-2">
                <div className="flex gap-x-2">
                  <span className="text-sm font-semibold">
                    {moment(samlFlow.initiateTime.toDate()).format()}
                  </span>
                  <span className="text-sm">Initiated SAML Flow</span>
                </div>

                <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
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
                  <span className="text-sm font-semibold">
                    {moment(samlFlow.receiveAssertionTime.toDate()).format()}
                  </span>
                  <span className="text-sm">Received SAML Assertion</span>
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

                <div className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm max-w-full overflow-auto">
                  {samlFlow.appRedirectUrl}
                </div>
              </div>
            </Card>
          )}
          {samlFlow?.redeemTime && (
            <Card>
              <div className="p-6 space-y-2">
                <div className="flex gap-x-2">
                  <span className="text-sm font-semibold">
                    {moment(samlFlow.redeemTime.toDate()).format()}
                  </span>
                  <span className="text-sm">Redeemed SAML Access Code</span>
                </div>
              </div>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
