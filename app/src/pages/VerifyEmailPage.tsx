import { useMutation } from "@connectrpc/connect-query/dist/esm";
import { signIn } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import React, { useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { useQuery } from "@connectrpc/connect-query";
import { setSessionToken } from "@/auth";
import { useNavigate } from "react-router";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export function VerifyEmailPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const emailVerifyToken = searchParams.get("t");
  const { data } = useQuery(signIn, {
    emailVerifyToken: emailVerifyToken ?? undefined,
  });

  useEffect(() => {
    if (data?.sessionToken) {
      setSessionToken(data.sessionToken);
      navigate("/");
    }
  }, [data, navigate]);

  return (
    <div className="h-screen flex items-center">
      <Card className="max-w-[400px] mx-auto">
        <CardHeader>
          <CardTitle>Verifying your email</CardTitle>
        </CardHeader>
        <CardContent className="text-sm">
          You'll be automatically redirected when this completes.
        </CardContent>
      </Card>
    </div>
  );
}
