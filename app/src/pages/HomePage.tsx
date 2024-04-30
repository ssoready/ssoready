import React from "react";
import { useQuery } from "@connectrpc/connect-query";
import { listEnvironments } from "../gen/ssoready/v1/ssoready-SSOReadyService_connectquery";

export function HomePage() {
  const { data: listEnvsRes } = useQuery(listEnvironments, {});
  return <h1>{JSON.stringify(listEnvsRes)}</h1>;
}
