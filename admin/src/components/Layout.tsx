import React, { ReactNode } from "react";
import {
  Disclosure,
  DisclosureButton,
  DisclosurePanel,
  Menu,
  MenuButton,
  MenuItem,
  MenuItems,
} from "@headlessui/react";
import { Bars3Icon, BellIcon, XMarkIcon } from "@heroicons/react/24/outline";
import { Outlet, useLocation } from "react-router";
import { Link } from "react-router-dom";
import { useQuery } from "@connectrpc/connect-query";
import { adminWhoami } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";

export function Layout() {
  const { data: whoami } = useQuery(adminWhoami, {});

  return (
    <>
      {/*
        This example requires updating your template:

        ```
        <html class="h-full">
        <body class="h-full">
        ```
      */}
      <div className="min-h-full">
        <nav className="border-b border-gray-200 bg-white">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="flex h-16 justify-between">
              <div className="flex">
                <Link
                  to="/"
                  className="flex flex-shrink-0 items-center text-sm"
                >
                  {whoami?.adminLogoUrl && (
                    <img
                      className="h-8 w-8 mr-4"
                      src={whoami.adminLogoUrl}
                      alt=""
                    />
                  )}

                  {whoami?.adminApplicationName
                    ? whoami.adminApplicationName
                    : "Settings Panel"}
                </Link>
              </div>

              {whoami?.adminReturnUrl && (
                <Link
                  to={whoami.adminReturnUrl}
                  className="flex items-center text-sm ml-auto"
                >
                  {whoami?.adminApplicationName
                    ? `Back to ${whoami.adminApplicationName}`
                    : "Exit"}
                </Link>
              )}
            </div>
          </div>
        </nav>

        <div>
          <Outlet />
        </div>
      </div>
    </>
  );
}

export function LayoutMain({ children }: { children?: ReactNode }) {
  return (
    <main>
      <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        {children}
      </div>
    </main>
  );
}
