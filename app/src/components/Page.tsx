import React from "react";
import { Outlet } from "react-router";
import { EnvironmentSwitcher } from "@/components/EnvironmentSwitcher";

export function Page() {
  return (
    <div>
      <div className="h-full border-r w-72 fixed bg-white p-2">
        <EnvironmentSwitcher />
      </div>
      <div className="ml-72 p-8">
        <Outlet />
      </div>
    </div>
  );
}
