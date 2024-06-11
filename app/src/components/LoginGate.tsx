import { useQuery } from "@connectrpc/connect-query";
import { whoami } from "../gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { ReactNode, useEffect } from "react";
import { Outlet, useNavigate } from "react-router";
import React from "react";
import * as Sentry from "@sentry/react";
import { analytics } from "@/analytics";

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
      navigate("/login", {
        replace: true,
      });
    }

    if (data) {
      Sentry.setUser({
        id: data.appUserId,
        email: data.email,
        username: data.displayName,
      });

      analytics.identify(data.appUserId);
    }
  }, [data, error, navigate]);

  return data ? <Outlet /> : <></>;
}
