import React from "react";
import { useInfiniteQuery, useQuery } from "@connectrpc/connect-query";
import {
  getOnboardingState,
  listEnvironments,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import {
  ArrowRight,
  ArrowRightIcon,
  BookIcon,
  BookOpenTextIcon,
  PlusIcon,
} from "lucide-react";
import { DocsLink } from "@/components/DocsLink";
import { Title } from "@/components/Title";

export function HomePage() {
  const { data: onboardingState } = useQuery(getOnboardingState, {});
  const {
    data: listEnvironmentsResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    listEnvironments,
    { pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );
  return (
    <div className="flex flex-col gap-y-8">
      <Title title="Home" />

      {/*{onboardingState && onboardingState.onboardingEnvironmentId === "" ? (*/}
      {/*  <Card>*/}
      {/*    <CardHeader>*/}
      {/*      <CardTitle>Add SAML support this afternoon</CardTitle>*/}
      {/*      <CardDescription>*/}
      {/*        Our interactive onboarding flow shows you how SSOReady works, and*/}
      {/*        how you'll integrate it in your product.*/}
      {/*      </CardDescription>*/}
      {/*    </CardHeader>*/}

      {/*    <CardContent>*/}
      {/*      <div className="flex items-center gap-x-4">*/}
      {/*        <Button asChild>*/}
      {/*          <Link to="/onboarding">*/}
      {/*            Get started*/}
      {/*            <ArrowRightIcon className="h-4 w-4 ml-2" />*/}
      {/*          </Link>*/}
      {/*        </Button>*/}
      {/*        <span className="text-sm text-muted-foreground">or</span>*/}
      {/*        <Button variant="outline" asChild>*/}
      {/*          <a href="https://ssoready.com/docs">*/}
      {/*            <BookOpenTextIcon className="h-4 w-4 mr-2" />*/}
      {/*            Read the Docs*/}
      {/*          </a>*/}
      {/*        </Button>*/}
      {/*      </div>*/}
      {/*    </CardContent>*/}
      {/*  </Card>*/}
      {/*) : (*/}
      <Card>
        <CardHeader>
          <CardTitle>Welcome to SSOReady</CardTitle>
        </CardHeader>

        <CardContent>
          <div className="flex items-center gap-x-4">
            <Button variant="outline" asChild>
              <a href="https://ssoready.com/docs">
                <BookOpenTextIcon className="h-4 w-4 mr-2" />
                Read the Docs
              </a>
            </Button>
          </div>
        </CardContent>
      </Card>
      {/*)}*/}

      <Card>
        <CardHeader>
          <div className="lg:flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <CardTitle>
                Environments
                <DocsLink to="https://ssoready.com/docs/sso-ready-concepts/environments" />
              </CardTitle>

              <CardDescription className="mr-2">
                An environment corresponds to a deployment environment in your
                product, e.g. "Staging" or "Production".
              </CardDescription>
            </div>

            <Button className="mt-5 lg:mt-0" variant="outline" asChild>
              <Link to="/environments/new">
                <PlusIcon className="h-4 w-4 mr-2" />
                Create environment
              </Link>
            </Button>
          </div>
        </CardHeader>

        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Environment ID</TableHead>
                <TableHead>Display Name</TableHead>
                <TableHead>App Redirect URL</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {listEnvironmentsResponses?.pages
                .flatMap((page) => page.environments)
                .map((environment) => (
                  <TableRow key={environment.id}>
                    <TableCell>
                      <Link
                        to={`/environments/${environment.id}`}
                        className="underline underline-offset-4 decoration-muted-foreground"
                      >
                        {environment.id}
                      </Link>
                    </TableCell>
                    <TableCell>{environment.displayName}</TableCell>
                    <TableCell>{environment.redirectUrl}</TableCell>
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
