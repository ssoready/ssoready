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

  return (
    <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <p>Loading...</p>
      <p>
        If this message does not go away, you may have used this setup link
        previously. Setup links expire once you visit them.
      </p>
    </div>
  );
}
