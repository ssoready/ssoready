import { useNavigate, useParams } from "react-router";
import { disableQuery, useQuery } from "@connectrpc/connect-query";
import {
  getOrganization,
  getSAMLConnection,
  getSAMLFlow,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { useEffect } from "react";
import React from "react";

export function RedirectSAMLFlowPage() {
  const { samlFlowId } = useParams();
  const { data: samlFlow } = useQuery(getSAMLFlow, {
    id: samlFlowId,
  });
  const { data: samlConnection } = useQuery(
    getSAMLConnection,
    samlFlow
      ? {
          id: samlFlow.samlConnectionId,
        }
      : disableQuery,
  );
  const { data: organization } = useQuery(
    getOrganization,
    samlConnection
      ? {
          id: samlConnection.organizationId,
        }
      : disableQuery,
  );

  const navigate = useNavigate();
  useEffect(() => {
    if (organization && samlConnection) {
      navigate(
        `/environments/${organization.environmentId}/organizations/${organization.id}/saml-connections/${samlConnection.id}/flows/${samlFlowId}`,
      );
    }
  }, [navigate, organization, samlConnection, samlFlowId]);

  return <></>;
}
