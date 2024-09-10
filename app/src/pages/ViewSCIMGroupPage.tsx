import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import React from "react";
import { useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import {
  appGetSCIMGroup,
  appListSCIMUsers,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import hljs from "highlight.js/lib/core";
import { useInfiniteQuery } from "@connectrpc/connect-query";
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
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Title } from "@/components/Title";

export function ViewSCIMGroupPage() {
  const { environmentId, organizationId, scimDirectoryId, scimGroupId } =
    useParams();
  const { data: scimGroup } = useQuery(appGetSCIMGroup, {
    id: scimGroupId,
  });
  const {
    data: listSCIMUsersResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    appListSCIMUsers,
    { scimDirectoryId, scimGroupId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  return (
    <div className="flex flex-col gap-8">
      <Title title="SCIM Group" />

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
            <BreadcrumbLink asChild>
              <Link
                to={`/environments/${environmentId}/organizations/${organizationId}/scim-directories/${scimDirectoryId}`}
              >
                {scimDirectoryId}
              </Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{scimGroupId}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>SCIM Group</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {scimGroupId}
                </span>
              </div>

              <CardDescription>
                A SCIM group is one of your customer's employees, synced in from
                their configured SCIM directory.
              </CardDescription>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-4 gap-y-2">
            <div className="text-sm col-span-1 text-muted-foreground">
              Display Name
            </div>
            <div className="text-sm col-span-3">{scimGroup?.displayName}</div>
            <div className="text-sm col-span-1 text-muted-foreground">
              Deleted
            </div>
            <div className="text-sm col-span-3">
              {scimGroup?.deleted ? "Yes" : "No"}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Group Details</CardTitle>
          <CardDescription>
            Your customers, depending on their Identity Provider vendor and
            configuration, will likely have a different set of{" "}
            <code className="font-mono text-xs text-foreground bg-gray-100 px-1 py-0.5 rounded-sm">
              attributes
            </code>{" "}
            on their groups.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <div className="text-xs font-mono bg-gray-100 py-2 px-2 rounded-sm max-w-full overflow-auto">
            <code>
              <code>
                {scimGroup && (
                  <pre
                    dangerouslySetInnerHTML={{
                      __html: hljs.highlight(
                        JSON.stringify(scimGroup, null, 4),
                        {
                          language: "json",
                        },
                      ).value,
                    }}
                  />
                )}
              </code>
            </code>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>SCIM Users</CardTitle>
          <CardDescription>Users belonging to this SCIM group.</CardDescription>
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
    </div>
  );
}
