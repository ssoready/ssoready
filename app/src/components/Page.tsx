import React from "react";
import { Outlet } from "react-router";
import { EnvironmentSwitcher } from "@/components/EnvironmentSwitcher";
import { Building2, LayoutGrid } from "lucide-react";

export function Page() {
  return (
    <div>
      <div className="h-full border-r w-72 fixed bg-white p-2">
        <EnvironmentSwitcher />

        <div className="m-2">
          <div className="flex gap-2 items-center p-2 hover:bg-gray-100 rounded-sm text-sm">
            <LayoutGrid className="h-4 w-4" />
            <span>Overview</span>
          </div>

          <div className="flex gap-2 items-center p-2 hover:bg-gray-100 rounded-sm text-sm">
            <Building2 className="h-4 w-4" />
            <span>Organizations</span>
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
