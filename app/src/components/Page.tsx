import React from "react";
import { Outlet } from "react-router";

export function Page() {
  return (
    <div>
      <div className="h-full border-r w-72 fixed bg-white">sidebar</div>
      <div className="ml-72 p-8">
        <Outlet />
      </div>
    </div>
  );
}
