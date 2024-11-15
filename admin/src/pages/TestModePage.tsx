import React, { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";

export function TestModePage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const idp = searchParams.get("idp")!;
  const samlConnectionId = searchParams.get("saml_connection_id")!;
  const email = searchParams.get("email")!;
  const attributes = searchParams.get("attributes")!;

  useEffect(() => {
    if (idp === "okta") {
      const search = new URLSearchParams({ email, attributes });
      navigate(
        `/saml/saml-connections/${samlConnectionId}/setup/okta-test-success?${search.toString()}`,
        {
          replace: true,
        },
      );
    }
  }, [navigate, idp, samlConnectionId, email, attributes]);

  return <></>;
}
