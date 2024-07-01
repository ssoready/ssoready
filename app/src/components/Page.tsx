import React from "react";
import { Outlet, useParams } from "react-router";

import { MobNav } from "./MobNav";
import { PageComp } from "./PageComp";

export function Page() {
  return (
    <div>
      <div className="hidden md:flex ">
        <PageComp />
      </div>

      <MobNav />
      <div className="md:ml-72 p-4 md:p-8">
        <div className="mx-auto max-w-6xl">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
