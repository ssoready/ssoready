import React from "react";
import { useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import { listOrganizations } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Link } from "react-router-dom";

export function ViewEnvironmentPage() {
  const { environmentId } = useParams();
  const { data: listOrgsRes } = useQuery(listOrganizations, {
    environmentId,
  });

  return (
    <>
      <Card>
        <CardHeader>{environmentId}</CardHeader>
        <CardContent>
          <div className="font-semibold">Environment Details</div>
          <div className="flex justify-between">
            <div>Redirect URL</div>
            <div>...</div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Organizations</CardTitle>
          <CardDescription>
            Organizations within this environment.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <Table>
            <TableHeader>
              <TableHead>Organization ID</TableHead>
            </TableHeader>
            <TableBody>
              {listOrgsRes?.organizations?.map((org) => (
                <TableRow>
                  <TableCell>
                    <Link
                      to={`/environments/${environmentId}/organizations/${org?.id}`}
                    >
                      {org.id}
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
