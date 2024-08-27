import React, { useCallback } from "react";
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  adminCreateSAMLConnection,
  adminListSAMLConnections,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { LayoutHeader, LayoutMain } from "@/components/Layout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Link, useNavigate } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CableIcon, CogIcon, PlusIcon } from "lucide-react";
import { toast } from "sonner";

export function ListSAMLConnectionsPage() {
  const {
    data: listSAMLConnectionResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    adminListSAMLConnections,
    { pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  const navigate = useNavigate();
  const createSAMLConnectionMutation = useMutation(adminCreateSAMLConnection);
  const handleCreateSAMLConnection = useCallback(async () => {
    const samlConnection = await createSAMLConnectionMutation.mutateAsync({
      samlConnection: {
        primary:
          listSAMLConnectionResponses?.pages.flatMap(
            (page) => page.samlConnections,
          ).length === 0,
      },
    });

    navigate(
      `/saml/saml-connections/${samlConnection.samlConnection!.id}/setup`,
    );
  }, [
    listSAMLConnectionResponses,
    createSAMLConnectionMutation.mutateAsync,
    navigate,
  ]);

  const samlConnections = listSAMLConnectionResponses?.pages.flatMap(
    (page) => page.samlConnections,
  );

  return (
    <>
      <LayoutMain>
        <Card>
          <CardHeader>
            <div className="flex justify-between items-center">
              <div className="flex flex-col space-y-1.5">
                <CardTitle>SAML Connections</CardTitle>
              </div>

              <Button variant="outline" onClick={handleCreateSAMLConnection}>
                <PlusIcon className="mr-2 h-4 w-4" />
                Create SAML connection
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {(samlConnections?.length ?? 0) > 0 ? (
              <>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>SAML Connection ID</TableHead>
                      <TableHead>IDP Redirect URL</TableHead>
                      <TableHead>IDP Entity ID</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {samlConnections?.map((samlConn) => (
                      <TableRow key={samlConn.id}>
                        <TableCell>
                          {samlConn.id}
                          {samlConn.primary && (
                            <Badge className="ml-2">Primary</Badge>
                          )}
                        </TableCell>
                        <TableCell className="max-w-[300px] truncate">
                          {samlConn.idpRedirectUrl}
                        </TableCell>
                        <TableCell className="max-w-[300px] truncate">
                          {samlConn.idpEntityId}
                        </TableCell>
                        <TableCell>
                          <Link
                            to={`/saml/saml-connections/${samlConn.id}/setup`}
                            className="underline underline-offset-4 decoration-muted-foreground"
                          >
                            <CogIcon className="inline h-4 w-4 mr-2" />
                            Configure
                          </Link>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>

                {hasNextPage && (
                  <Button variant="secondary" onClick={() => fetchNextPage()}>
                    Load more
                  </Button>
                )}
              </>
            ) : (
              <div className="mt-4 flex flex-col items-center">
                <CableIcon className="h-10 w-10 text-muted-foreground" />
                <div className="mt-2 text-sm font-semibold">
                  No SAML connections
                </div>
                <div className="mt-2 text-sm text-muted-foreground">
                  Get started by creating a new SAML connection.
                </div>

                <Button className="mt-4" onClick={handleCreateSAMLConnection}>
                  <PlusIcon className="mr-2 h-4 w-4" />
                  Create SAML connection
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </LayoutMain>
    </>
  );
}
