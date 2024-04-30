import React from "react";
import { GoogleLogin, GoogleOAuthProvider } from "@react-oauth/google";
import { useMutation } from "@connectrpc/connect-query";
import { signIn } from "../gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { setSessionToken } from "../auth";

export function LoginPage() {
  const signInMutation = useMutation(signIn);

  return (
    <GoogleOAuthProvider
      clientId={
        "171906208332-m8dg2p6av2f0aa7lliaj6oo0grct57p1.apps.googleusercontent.com"
      }
    >
      <GoogleLogin
        onSuccess={async (credentialResponse) => {
          const { sessionToken } = await signInMutation.mutateAsync({
            googleCredential: credentialResponse.credential,
          });
          setSessionToken(sessionToken);
        }}
      />
    </GoogleOAuthProvider>
  );
}
