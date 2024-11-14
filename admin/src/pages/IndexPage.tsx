import { LayoutMain } from "@/components/Layout";
import React, { useCallback } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  CableIcon,
  CogIcon,
  KeyIcon,
  Plus,
  PlusIcon,
  UsersIcon,
} from "lucide-react";
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  adminCreateSAMLConnection,
  adminCreateSCIMDirectory,
  adminListSAMLConnections,
  adminListSCIMDirectories,
  adminWhoami,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { useTitle } from "@/useTitle";
import { Helmet } from "react-helmet";
import { toast } from "sonner";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export function IndexPage() {
  const { data: whoami } = useQuery(adminWhoami, {});
  const title = useTitle("Home");
  const navigate = useNavigate();

  const {
    data: listSAMLConnectionResponses,
    fetchNextPage: fetchNextSAMLConnectionPage,
    hasNextPage: hasNextSAMLConnectionPage,
  } = useInfiniteQuery(
    adminListSAMLConnections,
    { pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

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

  const {
    data: listSCIMDirectoriesPages,
    fetchNextPage: fetchNextSCIMDirectoryPage,
    hasNextPage: hasNextSCIMDirectoryPage,
  } = useInfiniteQuery(
    adminListSCIMDirectories,
    { pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  const createSCIMDirectoryMutation = useMutation(adminCreateSCIMDirectory);
  const handleCreateSCIMDirectory = useCallback(async () => {
    const { scimDirectory } = await createSCIMDirectoryMutation.mutateAsync({
      scimDirectory: {
        primary:
          listSCIMDirectoriesPages?.pages.flatMap(
            (page) => page.scimDirectories,
          ).length === 0,
      },
    });

    toast("SCIM Directory has been created.");
    navigate(`/scim/scim-directories/${scimDirectory!.id}`);
  }, [createSCIMDirectoryMutation, navigate]);

  const scimDirectories = listSCIMDirectoriesPages?.pages.flatMap(
    (page) => page.scimDirectories,
  );

  return (
    <LayoutMain>
      <Helmet>
        <title>{title}</title>
      </Helmet>

      <div className="flex flex-col gap-y-8">
        {whoami?.canManageSaml && (
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <div className="flex flex-col space-y-1.5">
                  <CardTitle>SAML Connections</CardTitle>
                  <CardDescription>
                    SAML connections allow you to log in using enterprise
                    single-sign on.
                  </CardDescription>
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
                            <Link
                              to={`/saml/saml-connections/${samlConn.id}`}
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

                  {hasNextSAMLConnectionPage && (
                    <Button
                      variant="secondary"
                      onClick={() => fetchNextSAMLConnectionPage()}
                    >
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
        )}

        {whoami?.canManageScim && (
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <div className="flex flex-col space-y-1.5">
                  <CardTitle>SCIM Directories</CardTitle>
                  <CardDescription>
                    SCIM directories allow you to provision and deprovision
                    users using enterprise directory sync.
                  </CardDescription>
                </div>

                <Button variant="outline" onClick={handleCreateSCIMDirectory}>
                  <Plus className="mr-2 h-4 w-4" />
                  Create SCIM directory
                </Button>
              </div>
            </CardHeader>

            <CardContent>
              {(scimDirectories?.length ?? 0) > 0 ? (
                <>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>SCIM Directory ID</TableHead>
                        <TableHead>SCIM Base URL</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {listSCIMDirectoriesPages?.pages
                        .flatMap((page) => page.scimDirectories)
                        ?.map((scimDirectory) => (
                          <TableRow key={scimDirectory.id}>
                            <TableCell>
                              <Link
                                to={`/scim/scim-directories/${scimDirectory.id}`}
                                className="underline underline-offset-4 decoration-muted-foreground"
                              >
                                {scimDirectory.id}
                              </Link>
                              {scimDirectory.primary && (
                                <Badge className="ml-2">Primary</Badge>
                              )}
                            </TableCell>
                            <TableCell className="max-w-[300px] truncate">
                              {scimDirectory.scimBaseUrl}
                            </TableCell>
                          </TableRow>
                        ))}
                    </TableBody>
                  </Table>

                  {hasNextSCIMDirectoryPage && (
                    <Button
                      variant="secondary"
                      onClick={() => fetchNextSCIMDirectoryPage()}
                    >
                      Load more
                    </Button>
                  )}
                </>
              ) : (
                <div className="mt-4 flex flex-col items-center">
                  <CableIcon className="h-10 w-10 text-muted-foreground" />
                  <div className="mt-2 text-sm font-semibold">
                    No SCIM directories
                  </div>
                  <div className="mt-2 text-sm text-muted-foreground">
                    Get started by creating a new SCIM directory.
                  </div>

                  <Button className="mt-4" onClick={handleCreateSCIMDirectory}>
                    <PlusIcon className="mr-2 h-4 w-4" />
                    Create SCIM directory
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </LayoutMain>
  );
}
