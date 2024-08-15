import React from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  getAppOrganization,
  getStripeBillingPortalURL,
  getStripeCheckoutURL,
  listAppUsers,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { UserIcon } from "lucide-react";
import { Button } from "@/components/ui/button";

export function SettingsPage() {
  const { data: appOrganization } = useQuery(getAppOrganization, {});
  const { data: appUsers } = useQuery(listAppUsers, {});

  const getStripeBillingPortalURLMutation = useMutation(
    getStripeBillingPortalURL,
  );
  const handleClickManageBilling = async () => {
    const { url } = await getStripeBillingPortalURLMutation.mutateAsync({});
    window.location.href = url;
  };

  const getStripeCheckoutURLMutation = useMutation(getStripeCheckoutURL);
  const handleClickUpsell = async () => {
    const { url } = await getStripeCheckoutURLMutation.mutateAsync({});
    window.location.href = url;
  };

  return (
    <div>
      <Card>
        <CardHeader>
          <CardTitle>Billing</CardTitle>
          <CardDescription>
            Manage your payment methods and subscription.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {appOrganization && appOrganization.hasStripeCustomerId && (
            <>
              <Button variant="outline" onClick={handleClickManageBilling}>
                Manage Billing
              </Button>
            </>
          )}

          {appOrganization && !appOrganization.hasStripeCustomerId && (
            <>
              <p className="text-sm text-foreground">
                SSOReady Pro gives you access to the Management API, which lets
                you programmatically manage SSOReady organizations, SAML
                connections, SCIM directories, and customer self-serve setup
                URLs.
              </p>

              <Button className="mt-4" onClick={handleClickUpsell}>
                Upgrade to Pro
              </Button>
            </>
          )}
        </CardContent>
      </Card>

      <Card className="mt-8">
        <CardHeader>
          <CardTitle>Team Members</CardTitle>
          {appOrganization?.googleHostedDomain && (
            <CardDescription>
              Your coworkers can join this team automatically by logging in with
              their{" "}
              <span className="font-semibold">
                {appOrganization.googleHostedDomain}
              </span>{" "}
              Google account.
            </CardDescription>
          )}
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-y-4">
            {appUsers?.appUsers.map((appUser) => (
              <div key={appUser.id}>
                <div className="flex items-center gap-3">
                  <Avatar className="h-9 w-9">
                    <AvatarFallback>
                      <UserIcon />
                    </AvatarFallback>
                  </Avatar>
                  <div className="grid gap-0.5 text-xs">
                    <div className="font-medium truncate">
                      {appUser.displayName}
                    </div>
                    <div className="text-gray-400 truncate">
                      {appUser.email}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
