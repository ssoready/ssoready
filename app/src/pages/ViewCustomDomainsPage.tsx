import React from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import { getEnvironmentCustomDomainSettings } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Badge } from "@/components/ui/badge";

export function ViewCustomDomainsPage() {
  const { environmentId } = useParams();
  const { data: customDomainsSettings } = useQuery(
    getEnvironmentCustomDomainSettings,
    {
      environmentId,
    },
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          Custom Domain for{" "}
          <code className="text-xl bg-gray-100 py-1 px-2 rounded-sm">
            auth.ssoready.com
          </code>
        </CardTitle>
        <CardDescription>foo bar</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-4 gap-y-2">
          <div className="text-sm text-muted-foreground">Record Type</div>
          <div className="col-span-3 flex items-center text-sm">CNAME</div>

          <div className="text-sm text-muted-foreground">Record Name</div>
          <div className="col-span-3 flex items-center text-sm font-mono">
            {customDomainsSettings?.customAuthDomain}
          </div>

          <div className="text-sm text-muted-foreground">Record Value</div>
          <div className="col-span-3 flex items-center text-sm font-mono">
            {customDomainsSettings?.customAuthDomainCnameValue}
          </div>

          <div className="text-sm text-muted-foreground">Status</div>
          <div className="col-span-3 flex items-center">
            {customDomainsSettings?.customAuthDomainConfigured ? (
              <Badge
                variant="outline"
                className="inline-flex items-center gap-x-2"
              >
                <span className="relative flex h-3 w-3">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                </span>
                <span>Ready</span>
              </Badge>
            ) : (
              <Badge variant="outline">Pending</Badge>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
