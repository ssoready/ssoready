import React from "react";
import { useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import {
  getOrganization,
  getSAMLConnection,
  listSAMLConnections,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export function ViewSAMLConnectionPage() {
  const { samlConnectionId } = useParams();
  const { data: samlConnection } = useQuery(getSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>{samlConnection?.id}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="font-semibold">SAML Connection Details</div>
          <div className="flex justify-between">
            <div>Identity Provider Entity ID</div>
            <div>{samlConnection?.idpEntityId}</div>
          </div>
          <div className="flex justify-between">
            <div>Identity Provider Redirect URL</div>
            <div>{samlConnection?.idpRedirectUrl}</div>
          </div>
          <div className="flex justify-between">
            <div>Identity Provider x509 Certificate</div>
            <div>{samlConnection?.idpX509Certificate}</div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Identity Provider X.509 Certificate</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="bg-black p-6 rounded-md text-white text-sm">
            <pre>
              <code>{samlConnection?.idpX509Certificate}</code>
            </pre>
          </div>
        </CardContent>
      </Card>
    </>
  );
}
