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
  appGetSCIMUser,
  appListSCIMGroups,
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
import { Dock } from "lucide-react";
import { DocsLink } from "@/components/DocsLink";
import { InfoTooltip } from "@/components/InfoTooltip";

export function ViewSCIMUserPage() {
  const { environmentId, organizationId, scimDirectoryId, scimUserId } =
    useParams();
  const { data: scimUser } = useQuery(appGetSCIMUser, {
    id: scimUserId,
  });
  const {
    data: listSCIMGroupsResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    appListSCIMGroups,
    { scimDirectoryId, scimUserId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  return (
    <div className="flex flex-col gap-8">
      <Title title="SCIM User" />

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
            <BreadcrumbPage>{scimUserId}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>SCIM User</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {scimUserId}
                </span>
              </div>

              <CardDescription>
                A SCIM user is one of your customer's employees, synced in from
                their configured SCIM directory.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/scim-users" />
              </CardDescription>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-5 gap-y-2">
            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              Email
              <InfoTooltip>
                Your end user's email address.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/scim-users#email" />
              </InfoTooltip>
            </div>
            <div className="text-sm col-span-3">{scimUser?.email}</div>
            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              Deleted
              <InfoTooltip>
                Whether the user has been deleted or deprovisioned.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/scim-users#deleted" />
              </InfoTooltip>
            </div>
            <div className="text-sm col-span-3">
              {scimUser?.deleted ? "Yes" : "No"}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>
            User Details
            <DocsLink to="https://ssoready.com/docs/ssoready-concepts/scim-users#attributes" />
          </CardTitle>
          <CardDescription>
            Your customers, depending on their Identity Provider vendor and
            configuration, will likely have a different set of{" "}
            <code className="font-mono text-xs text-foreground bg-gray-100 px-1 py-0.5 rounded-sm">
              attributes
            </code>{" "}
            on their users.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <div className="text-xs font-mono bg-gray-100 py-2 px-2 rounded-sm max-w-full overflow-auto">
            <code>
              <code>
                {scimUser && (
                  <pre
                    dangerouslySetInnerHTML={{
                      __html: hljs.highlight(
                        JSON.stringify(scimUser, null, 4),
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
          <CardTitle>
            SCIM Groups
            <DocsLink to="https://ssoready.com/docs/ssoready-concepts/scim-users#groups" />
          </CardTitle>
          <CardDescription>
            SCIM groups this SCIM user belongs to.
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
    </div>
  );
}
