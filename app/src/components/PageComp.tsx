import React from "react";
import { Outlet, useParams } from "react-router";
import { EnvironmentSwitcher } from "@/components/EnvironmentSwitcher";
import {
  KeyRoundIcon,
  CalendarIcon,
  LayoutGrid,
  MailIcon,
  PhoneIcon,
  UserIcon,
} from "lucide-react";
import { Link } from "react-router-dom";
import { useQuery } from "@connectrpc/connect-query";
import { whoami } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

export function PageComp() {
  const { environmentId } = useParams();
  const { data: whoamiData } = useQuery(whoami, {});

  return (
    <div>
      <div className="flex flex-col h-full border-r md:w-72 fixed bg-white  justify-between">
        <div className="md:p-2">
          <EnvironmentSwitcher />

          <div className="m-2">
            {environmentId && (
              <Link
                to={`/environments/${environmentId}`}
                className="flex gap-2 items-center p-2 hover:bg-gray-100 rounded-sm text-sm"
              >
                <LayoutGrid className="h-4 w-4" />
                <span>Overview</span>
              </Link>
            )}

            {environmentId && (
              <Link
                to={`/environments/${environmentId}/api-keys`}
                className="flex gap-2 items-center md:p-2 hover:bg-gray-100 rounded-sm text-sm"
              >
                <KeyRoundIcon className="h-4 w-4" />
                <span>API Keys</span>
              </Link>
            )}
          </div>
        </div>

        <div>
          <div className="w-fit md:w-full border-t border-gray-200 mr-2 md:px-4 py-4">
            <div className="md:flex md:items-center md:gap-3 ">
              <div className="grid gap-0 break-after-right text-sm">
                <div className="font-medium">Call the SSOReady CTO</div>
                <div className="text-gray-400 text-xs">
                  Want to talk about SAML? You can talk to our CTO,{" "}
                  <a className="underline" href="https://github.com/ucarion">
                    Ulysse
                  </a>
                  , any time. He just loves this stuff.
                </div>
              </div>
              <img
                className="h-12 w-12 rounded-full"
                src="/ulysse.jpg"
                alt="CTO"
              />
            </div>

            <div className="mt-2 text-xs text-gray-600 flex flex-col gap-y-1">
              <div className="flex items-center">
                <PhoneIcon className="h-4 w-4 mr-2" />
                (510) 502 1557
              </div>
              <div className="flex items-center">
                <MailIcon className="h-4 w-4 mr-2" />
                ulysse.carion@ssoready.com
              </div>
              <div className="flex items-center">
                <CalendarIcon className="h-4 w-4 mr-2" />
                <a className="underline" href="https://cal.com/ucarion/30min">
                  Book a meeting
                </a>
              </div>
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
      </div>
    </div>
  );
}
