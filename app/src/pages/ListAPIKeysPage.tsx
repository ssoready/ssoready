import React, { useCallback, useState } from "react";
import {
  createAPIKey,
  createSAMLOAuthClient,
  listAPIKeys,
  listOrganizations,
  listSAMLOAuthClients,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { useParams } from "react-router";
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
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { Button } from "@/components/ui/button";
import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
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

export function ListAPIKeysPage() {
  return (
    <div className="flex flex-col gap-y-8">
      <ListAPIKeysCard />
      <ListOAuthClientsCard />
    </div>
  );
}

function ListAPIKeysCard() {
  const { environmentId } = useParams();
  const {
    data: listAPIKeysResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    listAPIKeys,
    { environmentId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  const [apiKeyId, setApiKeyId] = useState("");
  const [apiKeySecret, setApiKeySecret] = useState("");
  const [alertOpen, setAlertOpen] = useState(false);
  const createAPIKeyMutation = useMutation(createAPIKey);
  const queryClient = useQueryClient();
  const handleCreateAPIKey = useCallback(async () => {
    const apiKey = await createAPIKeyMutation.mutateAsync({
      apiKey: {
        environmentId,
      },
    });

    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey(listAPIKeys, { environmentId }),
    });

    setApiKeyId(apiKey.id);
    setApiKeySecret(apiKey.secretToken);
    setAlertOpen(true);
  }, [
    environmentId,
    createAPIKeyMutation,
    queryClient,
    setApiKeySecret,
    setAlertOpen,
  ]);

  return (
    <>
      <AlertDialog open={alertOpen} onOpenChange={setAlertOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>New API Key</AlertDialogTitle>
            <AlertDialogDescription>
              Your new API key has been created.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <Alert>
            <CircleAlert className="h-4 w-4" />
            <AlertTitle>Copy this secret</AlertTitle>
            <AlertDescription>
              Store this secret in a password or secrets manager. You can't
              retrieve it later.
            </AlertDescription>
          </Alert>

          <div className="text-sm font-medium leading-none">API Key Secret</div>

          <div className="text-xs font-mono bg-gray-100 py-2 px-4 rounded-sm border">
            {apiKeySecret}
          </div>

          <AlertDialogFooter>
            <AlertDialogAction asChild>
              <Link to={`/environments/${environmentId}/api-keys/${apiKeyId}`}>
                View API Key
              </Link>
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <CardTitle>API Keys</CardTitle>

              <CardDescription>
                An API key is how your application authenticates with SSOReady's
                SDKs or REST APIs. Most SSOReady users should use these over
                SAML OAuth Clients.
              </CardDescription>
            </div>

            <Button variant="outline" onClick={handleCreateAPIKey}>
              Create API Key
            </Button>
          </div>
        </CardHeader>

        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>SAML OAuth Client ID</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {listAPIKeysResponses?.pages
                .flatMap((page) => page.apiKeys)
                .map((apiKey) => (
                  <TableRow key={apiKey.id}>
                    <TableCell>
                      <Link
                        to={`/environments/${environmentId}/api-keys/${apiKey?.id}`}
                        className="underline underline-offset-4 decoration-muted-foreground"
                      >
                        {apiKey.id}
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
        </CardContent>
      </Card>
    </>
  );
}

function ListOAuthClientsCard() {
  const { environmentId } = useParams();
  const {
    data: listSAMLOAuthClientsResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    listSAMLOAuthClients,
    { environmentId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  const [samlOAuthClientId, setSAMLOAuthClientId] = useState("");
  const [samlOAuthClientSecret, setSAMLOAuthClientSecret] = useState("");
  const [samlOAuthClientAlertOpen, setSAMLOAuthClientAlertOpen] =
    useState(false);
  const createSAMLOAuthClientMutation = useMutation(createSAMLOAuthClient);
  const queryClient = useQueryClient();
  const handleCreateSAMLOAuthClient = useCallback(async () => {
    const samlOAuthClient = await createSAMLOAuthClientMutation.mutateAsync({
      samlOauthClient: {
        environmentId,
      },
    });

    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey(listSAMLOAuthClients, { environmentId }),
    });

    setSAMLOAuthClientId(samlOAuthClient.id);
    setSAMLOAuthClientSecret(samlOAuthClient.clientSecret);
    setSAMLOAuthClientAlertOpen(true);
  }, [
    environmentId,
    createSAMLOAuthClientMutation,
    queryClient,
    setSAMLOAuthClientSecret,
    setSAMLOAuthClientAlertOpen,
  ]);

  return (
    <>
      <AlertDialog
        open={samlOAuthClientAlertOpen}
        onOpenChange={setSAMLOAuthClientAlertOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>New SAML OAuth Client</AlertDialogTitle>
            <AlertDialogDescription>
              Your new SAML OAuth client has been created.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <Alert>
            <CircleAlert className="h-4 w-4" />
            <AlertTitle>Copy this client secret</AlertTitle>
            <AlertDescription>
              Store this client secret in a password or secrets manager. You
              can't retrieve it later.
            </AlertDescription>
          </Alert>

          <div className="text-sm font-medium leading-none">
            SAML OAuth Client Secret
          </div>

          <div className="text-xs font-mono bg-gray-100 py-2 px-4 rounded-sm border">
            {samlOAuthClientSecret}
          </div>

          <AlertDialogFooter>
            <AlertDialogAction asChild>
              <Link
                to={`/environments/${environmentId}/saml-oauth-clients/${samlOAuthClientId}`}
              >
                View SAML OAuth Client
              </Link>
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <CardTitle>SAML OAuth Clients</CardTitle>

              <CardDescription>
                A SAML OAuth client is a way to add SAML support to an app that
                exclusively uses OAuth to do user logins. Use these if
                interacting directly with SSOReady using an API Key isn't an
                option for you.
              </CardDescription>
            </div>

            <Button variant="outline" onClick={handleCreateSAMLOAuthClient}>
              Create SAML OAuth Client
            </Button>
          </div>
        </CardHeader>

        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>SAML OAuth Client ID</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {listSAMLOAuthClientsResponses?.pages
                .flatMap((page) => page.samlOauthClients)
                .map((samlOAuthClient) => (
                  <TableRow key={samlOAuthClient.id}>
                    <TableCell>
                      <Link
                        to={`/environments/${environmentId}/saml-oauth-clients/${samlOAuthClient?.id}`}
                        className="underline underline-offset-4 decoration-muted-foreground"
                      >
                        {samlOAuthClient.id}
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
        </CardContent>
      </Card>
    </>
  );
}
