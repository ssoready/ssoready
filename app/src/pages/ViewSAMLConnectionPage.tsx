import React from "react";
import { useMatch, useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import {
  getEnvironment,
  getOrganization,
  getSAMLConnection,
  listOrganizations,
  listSAMLConnections,
  listSAMLFlows,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import moment from "moment";

export function ViewSAMLConnectionPage() {
  const { environmentId, organizationId, samlConnectionId } = useParams();
  const { data: samlConnection } = useQuery(getSAMLConnection, {
    id: samlConnectionId,
  });
  const flowsPathMatch = useMatch(
    "/environments/:environmentId/organizations/:organizationId/saml-connections/:samlConnectionId/flows",
  );

  return (
    <div className="grid gap-8">
      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>SAML Connection</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {samlConnectionId}
                </span>
              </div>

              <CardDescription>
                A SAML connection is a link between your product and your
                customer's Identity Provider.
              </CardDescription>
            </div>

            {/*{environment && (*/}
            {/*    <EditEnvironmentAlertDialog environment={environment} />*/}
            {/*)}*/}
          </div>
        </CardHeader>
      </Card>

      <Tabs defaultValue={flowsPathMatch ? "flows" : "config"}>
        <TabsList>
          <TabsTrigger value="config">
            <Link
              to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}`}
            >
              Configuration
            </Link>
          </TabsTrigger>
          <TabsTrigger value="flows" asChild>
            <Link
              to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}/flows`}
            >
              Login Flows
            </Link>
          </TabsTrigger>
        </TabsList>
        <TabsContent value="config">
          <Card>
            <CardHeader>
              <CardTitle>Service Provider Configuration</CardTitle>
              <CardDescription>
                The configuration here is assigned automatically by SSOReady,
                and needs to be inputted into your customer's Identity Provider
                by their IT admin.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-4 gap-y-2 items-center">
                <div className="text-sm col-span-1 text-muted-foreground">
                  Entity ID
                </div>
                <div className="text-sm col-span-3">
                  {samlConnection?.spEntityId}
                </div>

                <div className="text-sm col-span-1 text-muted-foreground">
                  Assertion Consumer Service (ACS) URL
                </div>
                <div className="text-sm col-span-3">
                  {samlConnection?.spAcsUrl}
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="mt-4">
            <CardHeader>
              <CardTitle>Identity Provider Configuration</CardTitle>
              <CardDescription>
                The configuration here needs to be copied over from the
                customer's Identity Provider.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-4 gap-y-2 items-center">
                <div className="text-sm col-span-1 text-muted-foreground">
                  Entity ID
                </div>
                <div className="text-sm col-span-3">
                  {samlConnection?.idpEntityId}
                </div>
                <div className="text-sm col-span-1 text-muted-foreground">
                  Redirect URL
                </div>
                <div className="text-sm col-span-3">
                  {samlConnection?.idpRedirectUrl}
                </div>
              </div>

              <Collapsible className="mt-1.5">
                <CollapsibleTrigger className="text-sm text-muted-foreground">
                  Certificate (click to show)
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <div className="bg-black rounded-lg px-6 py-4 mt-4">
                    <code className="text-sm text-white">
                      <pre>{samlConnection?.idpCertificate}</pre>
                    </code>
                  </div>
                </CollapsibleContent>
              </Collapsible>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="flows">
          <ListLoginFlowsTabContent />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function ListLoginFlowsTabContent() {
  const { environmentId, organizationId, samlConnectionId } = useParams();
  const { data: listFlowsRes } = useQuery(listSAMLFlows, {
    samlConnectionId,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>SAML Login Flows</CardTitle>
        <CardDescription>
          SAML login flows from this connection are listed here.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>Timestamp</TableHead>
              <TableHead>User ID</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {listFlowsRes?.samlFlows?.map((flow) => (
              <TableRow key={flow.id}>
                <TableCell>{flow.id}</TableCell>
                <TableCell>
                  {moment(flow.createTime!.toDate()).format()}
                </TableCell>
                <TableCell>{flow.subjectIdpId}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
