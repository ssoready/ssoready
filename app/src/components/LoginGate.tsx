import { useQuery } from "@connectrpc/connect-query";
import { whoami } from "../gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { ReactNode, useEffect } from "react";
import { Outlet, useNavigate } from "react-router";
import React from "react";

export function LoginGate() {
  const { data, error } = useQuery(
    whoami,
    {},
    {
      retry: false,
    },
  );

  const navigate = useNavigate();
  useEffect(() => {
    if (error) {
      navigate("/login");
    }
  }, [error, navigate]);

  return data ? <Outlet /> : <></>;
}
