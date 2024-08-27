import React, { useCallback } from "react";
import { LayoutMain } from "@/components/Layout";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Link, useNavigate } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { useInfiniteQuery, useMutation } from "@connectrpc/connect-query";
import {
  adminCreateSCIMDirectory,
  adminListSCIMDirectories,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { toast } from "sonner";
import { Helmet } from "react-helmet";
import { useTitle } from "@/useTitle";

export function ListSCIMDirectoriesPage() {
  const title = useTitle("SCIM Directories");
  const {
    data: listSCIMDirectoriesPages,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    adminListSCIMDirectories,
    { pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  const createSCIMDirectoryMutation = useMutation(adminCreateSCIMDirectory);
  const navigate = useNavigate();
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

  return (
    <LayoutMain>
      <Helmet>
        <title>{title}</title>
      </Helmet>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <CardTitle className="flex gap-x-4 items-center">
                <span>SCIM Directories</span>
              </CardTitle>
            </div>

            <Button variant="outline" onClick={handleCreateSCIMDirectory}>
              <Plus className="mr-2 h-4 w-4" />
              Create SCIM directory
            </Button>
          </div>
        </CardHeader>

        <CardContent>
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

          {hasNextPage && (
            <Button variant="secondary" onClick={() => fetchNextPage()}>
              Load more
            </Button>
          )}
        </CardContent>
      </Card>
    </LayoutMain>
  );
}
