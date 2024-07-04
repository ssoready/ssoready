import React from "react";
import { Outlet } from "react-router";

import { MobSideBar } from "./MobSideBar";
import { PageComponent } from "./PageComponent";

export function Page() {
  return (
    <div>
      <MobSideBar />
      <div className="invisible md:h-full border-r w-72 fixed bg-white flex flex-col justify-between md:visible">
        <PageComponent />
      </div>
      <div className="md:ml-72 p-8">
        <div className="mx-auto max-w-6xl">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
