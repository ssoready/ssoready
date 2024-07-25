import React, { useEffect } from "react";
import { useMutation } from "@connectrpc/connect-query";
import { adminRedeemOneTimeToken } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { useNavigate, useSearchParams } from "react-router-dom";
import { setSessionToken } from "@/auth";

export function SetupPage() {
  const redeemOneTimeTokenMutation = useMutation(adminRedeemOneTimeToken);
  const [searchParams] = useSearchParams();
  const oneTimeToken = searchParams.get("one-time-token") ?? undefined;
  const navigate = useNavigate();

  useEffect(() => {
    (async () => {
      const { adminSessionToken } =
        await redeemOneTimeTokenMutation.mutateAsync({
          oneTimeToken,
        });

      setSessionToken(adminSessionToken);
      navigate("/");
    })();
  }, [oneTimeToken, redeemOneTimeTokenMutation.mutateAsync, navigate]);

  return <h1>setup page :)</h1>;
}
