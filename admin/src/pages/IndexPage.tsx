import { LayoutMain } from "@/components/Layout";
import React from "react";
import { Link } from "react-router-dom";
import { KeyIcon, UsersIcon } from "lucide-react";
import { useQuery } from "@connectrpc/connect-query";
import { adminWhoami } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";

export function IndexPage() {
  const { data: whoami } = useQuery(adminWhoami, {});

  return (
    <LayoutMain>
      <div className="container grid grid-cols-2 gap-6">
        {whoami?.canManageSaml && (
          <Link
            to="/saml"
            className="flex flex-col items-center justify-center gap-4 rounded-lg border border-input bg-card p-8 text-center transition-colors hover:bg-accent hover:text-accent-foreground focus:outline-none focus:ring-1 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
          >
            <div className="bg-muted rounded-full p-4">
              <KeyIcon className="h-8 w-8 text-muted-foreground" />
            </div>
            <div>
              <h3 className="text-2xl font-bold">Configure SAML</h3>
              <p className="text-muted-foreground">
                Set up single sign-on with your identity provider.
              </p>
            </div>
          </Link>
        )}
        {whoami?.canManageScim && (
          <Link
            to="/scim"
            className="flex flex-col items-center justify-center gap-4 rounded-lg border border-input bg-card p-8 text-center transition-colors hover:bg-accent hover:text-accent-foreground focus:outline-none focus:ring-1 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
          >
            <div className="bg-muted rounded-full p-4">
              <UsersIcon className="h-8 w-8 text-muted-foreground" />
            </div>
            <div>
              <h3 className="text-2xl font-bold">Configure SCIM</h3>
              <p className="text-muted-foreground">
                Automate user provisioning and deprovisioning.
              </p>
            </div>
          </Link>
        )}
      </div>
    </LayoutMain>
  );
}
