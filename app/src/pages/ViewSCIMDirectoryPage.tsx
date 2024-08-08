import React, { useState } from "react";
import { useMatch, useParams } from "react-router";
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  appListSCIMGroups,
  appListSCIMUsers,
  getSCIMDirectory,
  listSAMLFlows,
  rotateSCIMDirectoryBearerToken,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Link } from "react-router-dom";
import { DocsLink } from "@/components/DocsLink";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import moment from "moment/moment";
import { SAMLFlowStatus } from "@/gen/ssoready/v1/ssoready_pb";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { CircleAlert } from "lucide-react";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

export function ViewSCIMDirectoryPage() {
  const { environmentId, organizationId, scimDirectoryId } = useParams();
  const { data: scimDirectory } = useQuery(getSCIMDirectory, {
    id: scimDirectoryId,
  });

  const groupsPathMatch = useMatch(
    "/environments/:environmentId/organizations/:organizationId/scim-directories/:scimDirectoryId/groups",
  );

  const [bearerTokenAlertOpen, setBearerTokenAlertOpen] = useState(false);
  const [bearerToken, setBearerToken] = useState("");
  const rotateSCIMDirectoryBearerTokenMutation = useMutation(
    rotateSCIMDirectoryBearerToken,
  );
  const handleRotateBearerToken = async () => {
    const { bearerToken } =
      await rotateSCIMDirectoryBearerTokenMutation.mutateAsync({
        scimDirectoryId,
      });
    setBearerToken(bearerToken);
    setBearerTokenAlertOpen(true);
  };

  return (
    <>
      <AlertDialog
        open={bearerTokenAlertOpen}
        onOpenChange={setBearerTokenAlertOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Updated SCIM Bearer Token</AlertDialogTitle>
            <AlertDialogDescription>
              This SCIM directory's bearer token has been updated.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <Alert>
            <CircleAlert className="h-4 w-4" />
            <AlertTitle>Copy this bearer token</AlertTitle>
            <AlertDescription>
              Give this secret to your customer's IT admin. They will input it
              into their Identity Provider.
            </AlertDescription>
          </Alert>

          <div className="text-sm font-medium leading-none">
            SCIM Bearer Token
          </div>

          <div className="text-xs font-mono bg-gray-100 py-2 px-4 rounded-sm border">
            {bearerToken}
          </div>

          <AlertDialogFooter>
            <AlertDialogCancel>Close</AlertDialogCancel>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <div className="flex flex-col gap-y-8">
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbLink asChild>
                <Link
                  to={`/environments/${environmentId}/organizations/${organizationId}`}
                >
                  {organizationId}
                </Link>
              </BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage>{scimDirectoryId}</BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>

        <Card>
          <CardHeader>
            <div className="flex justify-between items-center">
              <div className="flex flex-col space-y-1.5">
                <div className="flex gap-4">
                  <CardTitle>SCIM Directory</CardTitle>

                  <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                    {scimDirectoryId}
                  </span>
                </div>

                <CardDescription>
                  A SCIM directory is a connection between your product and your
                  customer's employee directory.
                </CardDescription>
              </div>

              <div></div>
            </div>
          </CardHeader>

          <CardContent>
            <div className="grid grid-cols-4 gap-y-2">
              <div className="text-sm col-span-1 text-muted-foreground">
                Primary
              </div>
              <div className="text-sm col-span-3">
                {scimDirectory?.primary ? "Yes" : "No"}
              </div>
              <div className="text-sm col-span-1 text-muted-foreground">
                SCIM Base URL
              </div>
              <div className="text-sm col-span-3">
                {scimDirectory?.scimBaseUrl}
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>SCIM Authentication</CardTitle>
            <CardDescription>
              To set up SCIM, your customer's IT admin will need to configure a
              Bearer authentication token in their Identity Provider. You can
              generate these here. You can only view a bearer token immediately
              after creating it; SSOReady only stores a hash.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="outline" onClick={handleRotateBearerToken}>
              Generate new bearer token
            </Button>

            <p className="mt-4 text-sm text-foreground-muted">
              Generating a new bearer token will break any existing syncs with
              this SCIM directory.
            </p>
          </CardContent>
        </Card>

        <Tabs defaultValue={groupsPathMatch ? "groups" : "users"}>
          <TabsList>
            <TabsTrigger value="users" asChild>
              <Link
                to={`/environments/${environmentId}/organizations/${organizationId}/scim-directories/${scimDirectoryId}`}
              >
                Users
              </Link>
            </TabsTrigger>
            <TabsTrigger value="groups" asChild>
              <Link
                to={`/environments/${environmentId}/organizations/${organizationId}/scim-directories/${scimDirectoryId}/groups`}
              >
                Groups
              </Link>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="users">
            <UsersTabContent />
          </TabsContent>

          <TabsContent value="groups">
            <GroupsTabContent />
          </TabsContent>
        </Tabs>
      </div>
    </>
  );
}

function UsersTabContent() {
  const { environmentId, organizationId, scimDirectoryId } = useParams();
  const {
    data: listSCIMUsersResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    appListSCIMUsers,
    { scimDirectoryId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>SCIM Users</CardTitle>
        <CardDescription>
          Users belonging to this SCIM directory.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Deleted</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {listSCIMUsersResponses?.pages
              ?.flatMap((page) => page.scimUsers)
              ?.map((scimUser) => (
                <TableRow key={scimUser.id}>
                  <TableCell className="max-w-[200px] truncate">
                    <Link
                      to={`/environments/${environmentId}/organizations/${organizationId}/scim-directories/${scimDirectoryId}/users/${scimUser.id}`}
                      className="underline underline-offset-4 decoration-muted-foreground"
                    >
                      {scimUser.id}
                    </Link>
                  </TableCell>
                  <TableCell>{scimUser.email}</TableCell>
                  <TableCell>{scimUser.deleted ? "Yes" : "No"}</TableCell>
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
  );
}

function GroupsTabContent() {
  const { environmentId, organizationId, scimDirectoryId } = useParams();
  const {
    data: listSCIMGroupsResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    appListSCIMGroups,
    { scimDirectoryId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>SCIM Groups</CardTitle>
        <CardDescription>
          Groups belonging to this SCIM directory.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>Display Name</TableHead>
              <TableHead>Deleted</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {listSCIMGroupsResponses?.pages
              ?.flatMap((page) => page.scimGroups)
              ?.map((scimGroup) => (
                <TableRow key={scimGroup.id}>
                  <TableCell className="max-w-[200px] truncate">
                    <Link
                      to={`/environments/${environmentId}/organizations/${organizationId}/scim-directories/${scimDirectoryId}/groups/${scimGroup.id}`}
                      className="underline underline-offset-4 decoration-muted-foreground"
                    >
                      {scimGroup.id}
                    </Link>
                  </TableCell>
                  <TableCell>{scimGroup.displayName}</TableCell>
                  <TableCell>{scimGroup.deleted ? "Yes" : "No"}</TableCell>
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
  );
}
