import React from "react";
import { GoogleLogin, GoogleOAuthProvider } from "@react-oauth/google";

export function LoginPage() {
  return (
    <GoogleOAuthProvider
      clientId={
        "171906208332-m8dg2p6av2f0aa7lliaj6oo0grct57p1.apps.googleusercontent.com"
      }
    >
      <GoogleLogin
        onSuccess={(credentialResponse) =>
          console.log(credentialResponse.credential)
        }
      />
    </GoogleOAuthProvider>
  );
}
