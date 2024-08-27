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

function classNames(...classes: any[]) {
  return classes.filter(Boolean).join(" ");
}

export function Layout() {
  const location = useLocation();
  const { data: whoami } = useQuery(adminWhoami, {});

  const navigation = [
    ...(whoami?.canManageSaml
      ? [
          {
            name: "SAML Settings",
            href: "/saml",
            current: location.pathname.startsWith("/saml"),
          },
        ]
      : []),
    ...(whoami?.canManageScim
      ? [
          {
            name: "SCIM Settings",
            href: "/scim",
            current: location.pathname.startsWith("/scim"),
          },
        ]
      : []),
  ];

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
                <div className="hidden sm:-my-px sm:ml-6 sm:flex sm:space-x-8">
                  {navigation.map((item) => (
                    <a
                      key={item.name}
                      href={item.href}
                      aria-current={item.current ? "page" : undefined}
                      className={classNames(
                        item.current
                          ? "border-indigo-500 text-gray-900"
                          : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700",
                        "inline-flex items-center border-b-2 px-1 pt-1 text-sm font-medium",
                      )}
                    >
                      {item.name}
                    </a>
                  ))}
                </div>
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

export function LayoutHeader({ children }: { children?: ReactNode }) {
  return (
    <header>
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <h1 className="text-3xl font-bold leading-tight tracking-tight text-gray-900">
          {children}
        </h1>
      </div>
    </header>
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
