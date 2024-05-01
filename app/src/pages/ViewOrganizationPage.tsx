import React, { useCallback, useMemo } from "react";
import { useNavigate, useParams } from "react-router";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  createSAMLConnection,
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
import { Button } from "@/components/ui/button";
import { PlusCircle } from "lucide-react";

export function ViewOrganizationPage() {
  const { environmentId, organizationId } = useParams();
  const { data: organization } = useQuery(getOrganization, {
    id: organizationId,
  });
  const { data: listSAMLConnectionsRes } = useQuery(listSAMLConnections, {
    organizationId,
  });

  const navigate = useNavigate();
  const createSAMLConnectionMutation = useMutation(createSAMLConnection);
  const handleCreateSAMLConnection = useCallback(async () => {
    const samlConnection = await createSAMLConnectionMutation.mutateAsync({
      samlConnection: {
        organizationId,
      },
    });

    navigate(
      `/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnection.id}/edit`,
    );
  }, [organizationId]);

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
          <div className="flex justify-between">
            <div>
              <CardTitle>SAML Connections</CardTitle>
              <CardDescription>
                SAML Connections within this organization.
              </CardDescription>
            </div>

            <div>
              <Button
                size="sm"
                className="h-8 gap-1"
                onClick={handleCreateSAMLConnection}
              >
                <PlusCircle className="h-3.5 w-3.5" />
                <span className="sr-only sm:not-sr-only sm:whitespace-nowrap">
                  Add SAML Connection
                </span>
              </Button>
            </div>
          </div>
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
