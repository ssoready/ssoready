import React from "react";
import { useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import {
  getOrganization,
  listSAMLConnections,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Link } from "react-router-dom";

export function ViewOrganizationPage() {
  const { organizationId } = useParams();
  const { data: organization } = useQuery(getOrganization, {
    id: organizationId,
  });
  const { data: listSAMLConnectionsRes } = useQuery(listSAMLConnections, {
    organizationId,
  });

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>{organization?.id}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="font-semibold">Organization Details</div>
          <div className="flex justify-between">
            <div>External ID</div>
            <div>{organization?.externalId}</div>
          </div>
          <div className="flex justify-between">
            <div>Domains</div>
            <div>
              {organization?.domains?.map((domain, i) => (
                <Badge key={i}>{domain}</Badge>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>SAML Connections</CardTitle>
          <CardDescription>
            SAML Connections within this organization.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <Table>
            <TableHeader>
              <TableHead>SAML Connection ID</TableHead>
              <TableHead>IDP Redirect URL</TableHead>
              <TableHead>IDP Entity ID</TableHead>
            </TableHeader>
            <TableBody>
              {listSAMLConnectionsRes?.samlConnections?.map((samlConn) => (
                <TableRow>
                  <TableCell>
                    <Link
                      to={`/environments/${organization?.environmentId}/organizations/${organization?.id}/saml-connections/${samlConn.id}`}
                    >
                      {samlConn.id}
                    </Link>
                  </TableCell>
                  <TableCell>{samlConn.idpRedirectUrl}</TableCell>
                  <TableCell>{samlConn.idpEntityId}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </>
  );
}
