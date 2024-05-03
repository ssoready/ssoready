/**
 * v0 by Vercel.
 * @see https://v0.dev/t/qxfm7X1lFC9
 * Documentation: https://v0.dev/docs#integrating-generated-code-into-your-nextjs-app
 */
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
import { Box } from "lucide-react";
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
            <span className="font-medium">{currentEnv?.displayName}</span>
          </span>
          <ChevronDownIcon className="h-4 w-4" />
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
                  <CheckIcon className="h-5 w-5 text-green-500" />
                )}
              </div>
            </Link>
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <div className="flex items-center gap-2">
            <PlusIcon className="h-5 w-5 text-gray-500" />
            <span>Create new environment</span>
          </div>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function CheckIcon(props) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M20 6 9 17l-5-5" />
    </svg>
  );
}

function ChevronDownIcon(props) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="m6 9 6 6 6-6" />
    </svg>
  );
}

function PlusIcon(props) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M5 12h14" />
      <path d="M12 5v14" />
    </svg>
  );
}
