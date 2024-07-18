import { useMutation } from "@connectrpc/connect-query";
import { signIn } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import React, { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { setSessionToken } from "@/auth";
import { useNavigate } from "react-router";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ConnectError } from "@connectrpc/connect";

export function VerifyEmailPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const emailVerifyToken = searchParams.get("t");

  const signInMutation = useMutation(signIn);
  const [emailLoginsDisabled, setEmailLoginsDisabled] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const { sessionToken } = await signInMutation.mutateAsync({
          emailVerifyToken: emailVerifyToken ?? undefined,
        });
        setSessionToken(sessionToken);
        navigate("/");
      } catch (e) {
        if (
          e instanceof ConnectError &&
          e.message ===
            "[failed_precondition] app organization has email_logins_disabled"
        ) {
          setEmailLoginsDisabled(true);
        }
      }
    })();
  }, [emailVerifyToken, navigate, signInMutation.mutateAsync]);

  return (
    <div className="h-screen flex items-center">
      <Card className="max-w-[400px] mx-auto">
        {emailLoginsDisabled ? (
          <>
            <CardHeader>
              <CardTitle>Email verification disabled</CardTitle>
            </CardHeader>
            <CardContent className="text-sm">
              Your organization has email-based logins disabled.{" "}
              <Link className="underline" to="/login">
                Use another login method
              </Link>
              .
            </CardContent>
          </>
        ) : (
          <>
            <CardHeader>
              <CardTitle>Verifying your email</CardTitle>
            </CardHeader>
            <CardContent className="text-sm">
              You'll be automatically redirected when this completes.
            </CardContent>
          </>
        )}
      </Card>
    </div>
  );
}
