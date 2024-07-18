import React, { useCallback, useEffect } from "react";
import { useMutation } from "@connectrpc/connect-query";
import { signIn } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { CredentialResponse } from "@react-oauth/google";
import { setSessionToken } from "@/auth";
import { useNavigate, useParams } from "react-router";
import { useSearchParams } from "react-router-dom";

export function MicrosoftCallbackPage() {
  const signInMutation = useMutation(signIn);
  const [searchParams] = useSearchParams();
  const code = searchParams.get("code");
  const navigate = useNavigate();

  useEffect(() => {
    if (!code) {
      return;
    }

    (async () => {
      const { sessionToken } = await signInMutation.mutateAsync({
        microsoftCode: code,
      });
      setSessionToken(sessionToken);
      navigate("/");
    })();
  }, [code, navigate, signInMutation.mutateAsync]);

  return <></>;
}
