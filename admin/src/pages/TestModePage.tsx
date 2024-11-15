import React, { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";

export function TestModePage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const idp = searchParams.get("idp")!;
  const samlConnectionId = searchParams.get("saml_connection_id")!;
  const email = searchParams.get("email")!;
  const attributes = searchParams.get("attributes")!;

  const search = new URLSearchParams({ email, attributes }).toString();

  useEffect(() => {
    if (idp === "okta") {
      navigate(
        `/saml/saml-connections/${samlConnectionId}/setup/okta-test-success?${search}`,
        {
          replace: true,
        },
      );
    }

    if (idp === "google") {
      navigate(
        `/saml/saml-connections/${samlConnectionId}/setup/google-test-success?${search}`,
        {
          replace: true,
        },
      );
    }

    if (idp === "entra") {
      navigate(
        `/saml/saml-connections/${samlConnectionId}/setup/entra-test-success?${search}`,
        {
          replace: true,
        },
      );
    }
  }, [navigate, samlConnectionId, idp, search]);

  return <></>;
}
