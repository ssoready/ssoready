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
  listAppUsers,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { UserIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Title } from "@/components/Title";

export function SettingsPage() {
  const { data: appOrganization } = useQuery(getAppOrganization, {});
  const { data: appUsers } = useQuery(listAppUsers, {});

  return (
    <div>
      <Title title="Settings" />

      <Card>
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
