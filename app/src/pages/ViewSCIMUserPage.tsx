import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import moment from "moment/moment";
import React from "react";
import { useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import { appGetSCIMUser } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import hljs from "highlight.js/lib/core";

export function ViewSCIMUserPage() {
  const { environmentId, organizationId, scimDirectoryId, scimUserId } =
    useParams();
  const { data: scimUser } = useQuery(appGetSCIMUser, {
    id: scimUserId,
  });

  return (
    <div className="flex flex-col gap-8">
      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>SCIM User</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {scimUserId}
                </span>
              </div>

              <CardDescription>
                A SCIM user is one of your customer's employees, synced in from
                their configured SCIM directory.
              </CardDescription>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-4 gap-y-2">
            <div className="text-sm col-span-1 text-muted-foreground">
              Email
            </div>
            <div className="text-sm col-span-3">{scimUser?.email}</div>
            <div className="text-sm col-span-1 text-muted-foreground">
              Deleted
            </div>
            <div className="text-sm col-span-3">
              {scimUser?.deleted ? "Yes" : "No"}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>User Details</CardTitle>
          <CardDescription>
            Your customers, depending on their Identity Provider vendor and
            configuration, will likely have a different set of{" "}
            <code className="font-mono text-xs text-foreground bg-gray-100 px-1 py-0.5 rounded-sm">
              attributes
            </code>{" "}
            on their users.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <div className="text-xs font-mono bg-gray-100 py-2 px-2 rounded-sm max-w-full overflow-auto">
            <code>
              <code>
                {scimUser && (
                  <pre
                    dangerouslySetInnerHTML={{
                      __html: hljs.highlight(
                        JSON.stringify(scimUser, null, 4),
                        {
                          language: "json",
                        },
                      ).value,
                    }}
                  />
                )}
              </code>
            </code>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
