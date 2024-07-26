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
import { PlusIcon } from "lucide-react";
import { toast } from "sonner";

export function IndexPage() {
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

    toast("SAML Connection has been created.");
    navigate(`/saml-connections/${samlConnection.samlConnection!.id}`);
  }, [
    listSAMLConnectionResponses,
    createSAMLConnectionMutation.mutateAsync,
    navigate,
  ]);

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
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>SAML Connection ID</TableHead>
                  <TableHead>IDP Redirect URL</TableHead>
                  <TableHead>IDP Entity ID</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {listSAMLConnectionResponses?.pages
                  .flatMap((page) => page.samlConnections)
                  ?.map((samlConn) => (
                    <TableRow key={samlConn.id}>
                      <TableCell>
                        <Link
                          to={`/saml-connections/${samlConn.id}`}
                          className="underline underline-offset-4 decoration-muted-foreground"
                        >
                          {samlConn.id}
                        </Link>
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
                    </TableRow>
                  ))}
              </TableBody>
            </Table>

            {hasNextPage && (
              <Button variant="secondary" onClick={() => fetchNextPage()}>
                Load more
              </Button>
            )}
          </CardContent>
        </Card>
      </LayoutMain>
    </>
  );
}
