import React, { useEffect } from "react";
import { useMutation } from "@connectrpc/connect-query";
import { redeemStripeCheckout } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { useSearchParams } from "react-router-dom";
import { useNavigate } from "react-router";
import { setSessionToken } from "@/auth";
import { toast } from "sonner";

export function StripeCheckoutSuccessPage() {
  const redeemStripeCheckoutMutation = useMutation(redeemStripeCheckout);
  const [searchParams] = useSearchParams();
  const sessionId = searchParams.get("session_id");
  const navigate = useNavigate();

  useEffect(() => {
    if (!sessionId) {
      return;
    }

    (async () => {
      await redeemStripeCheckoutMutation.mutateAsync({
        stripeCheckoutSessionId: sessionId,
      });
      toast("Successfully updated your subscription plan.");
      navigate("/");
    })();
  }, [sessionId, navigate, redeemStripeCheckoutMutation.mutateAsync]);

  return <></>;
}
