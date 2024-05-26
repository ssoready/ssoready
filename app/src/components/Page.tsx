import React from "react";
import { Outlet, useParams } from "react-router";
import { EnvironmentSwitcher } from "@/components/EnvironmentSwitcher";
import { Building2, LayoutGrid, UserIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { useQuery } from "@connectrpc/connect-query";
import { whoami } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

export function Page() {
  const { environmentId } = useParams();
  const { data: whoamiData } = useQuery(whoami, {});

  return (
    <div>
      <div className="h-full border-r w-72 fixed bg-white flex flex-col justify-between">
        <div className="p-2">
          <EnvironmentSwitcher />

          <div className="m-2">
            <Link
              to={`/environments/${environmentId}`}
              className="flex gap-2 items-center p-2 hover:bg-gray-100 rounded-sm text-sm"
            >
              <LayoutGrid className="h-4 w-4" />
              <span>Overview</span>
            </Link>
          </div>
        </div>

        <div className="flex items-center gap-3 border-t border-gray-200 px-4 py-4">
          <Avatar className="h-9 w-9">
            <AvatarFallback>
              <UserIcon />
            </AvatarFallback>
          </Avatar>
          <div className="grid gap-0.5 text-xs">
            <div className="font-medium">{whoamiData?.displayName}</div>
            <div className="text-gray-400">{whoamiData?.email}</div>
          </div>
        </div>
      </div>
      <div className="ml-72 p-8">
        <div className="mx-auto max-w-6xl">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
