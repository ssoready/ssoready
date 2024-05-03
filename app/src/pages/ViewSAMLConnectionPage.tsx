import React from "react";
import { useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import {
  getOrganization,
  getSAMLConnection,
  listSAMLConnections,
  listSAMLFlows,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ChevronLeft } from "lucide-react";
import { Link } from "react-router-dom";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export function ViewSAMLConnectionPage() {
  const { environmentId, organizationId, samlConnectionId } = useParams();
  const { data: samlConnection } = useQuery(getSAMLConnection, {
    id: samlConnectionId,
  });
  const { data: listFlowsRes } = useQuery(listSAMLFlows, {
    samlConnectionId,
  });

  return (
    <>
      <div className="flex items-center gap-4">
        <Button variant="outline" size="icon" className="h-7 w-7">
          <ChevronLeft className="h-4 w-4" />
          <span className="sr-only">Back</span>
        </Button>
        <h1 className="flex-1 shrink-0 whitespace-nowrap text-xl font-semibold tracking-tight sm:grow-0">
          {samlConnectionId}
        </h1>
        {/*<Badge variant="outline" className="ml-auto sm:ml-0">*/}
        {/*  In stock*/}
        {/*</Badge>*/}
        <div className="hidden items-center gap-2 md:ml-auto md:flex">
          <Button size="sm" asChild>
            <Link
              to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}/edit`}
            >
              Edit
            </Link>
          </Button>
        </div>
      </div>

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
            <div>Service Provider Entity ID</div>
            <div>{samlConnection?.spEntityId}</div>
          </div>
          <div className="flex justify-between">
            <div>Service Provider ACS URL</div>
            <div>{samlConnection?.spAcsUrl}</div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Identity Provider Certificate</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="bg-black p-6 rounded-md text-white text-sm">
            <pre>
              <code>{samlConnection?.idpCertificate}</code>
            </pre>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Recent SAML Login Flows</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableHead>SAML Flow ID</TableHead>
            </TableHeader>
            <TableBody>
              {listFlowsRes?.samlFlows?.map((samlFlow) => (
                <TableRow>
                  <TableCell>
                    <Link
                      to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}/flows/${samlFlow.id}`}
                    >
                      {samlFlow.id}
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </>
  );
}
