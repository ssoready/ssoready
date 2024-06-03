import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import React from "react";
import { disableQuery, useQuery } from "@connectrpc/connect-query";
import {
  getEnvironment,
  listEnvironments,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Box, Check, ChevronDown, Plus } from "lucide-react";
import { Link } from "react-router-dom";
import { useParams } from "react-router";

export function EnvironmentSwitcher() {
  const { environmentId } = useParams();
  const { data: listEnvsRes } = useQuery(listEnvironments, {});
  const { data: currentEnv } = useQuery(
    getEnvironment,
    environmentId
      ? {
          id: environmentId,
        }
      : disableQuery,
  );

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          className="flex justify-between items-center w-full"
          variant="outline"
        >
          <span className="flex items-center gap-2">
            <Box className="h-4 w-4" />
            <span className="font-medium">
              {currentEnv ? (
                currentEnv.displayName
              ) : (
                <span className="font-normal text-gray-500">
                  No environment selected
                </span>
              )}
            </span>
          </span>
          <ChevronDown className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-[360px]">
        <DropdownMenuLabel>Environments</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {listEnvsRes?.environments.map((env) => (
          <DropdownMenuItem key={env.id} className="cursor-pointer" asChild>
            <Link to={`/environments/${env.id}`}>
              <div className="flex items-center justify-between w-full">
                <div>
                  <div className="font-medium">{env.displayName}</div>
                  <div className="text-sm text-gray-500 dark:text-gray-400 max-w-48 truncate">
                    {env.id}
                  </div>
                </div>
                {env.id === environmentId && (
                  <Check className="h-5 w-5 text-green-500" />
                )}
              </div>
            </Link>
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <Link className="flex items-center w-full" to="/environments/new">
            <Plus className="mr-2 h-4 w-4" />
            <span>New environment</span>
          </Link>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
