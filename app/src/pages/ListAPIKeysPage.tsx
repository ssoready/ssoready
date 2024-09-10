import React, { useCallback, useEffect, useState } from "react";
import {
  createAPIKey,
  createSAMLOAuthClient,
  getAppOrganization,
  getStripeCheckoutURL,
  listAPIKeys,
  listSAMLOAuthClients,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { useNavigate, useParams } from "react-router";
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
import { useInfiniteQuery, useQuery } from "@connectrpc/connect-query";
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
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { offset, useFloating, useTransitionStyles } from "@floating-ui/react";
import { CircleAlert, Plus, CopyIcon } from "lucide-react";
import { z } from "zod";
import { Environment } from "@/gen/ssoready/v1/ssoready_pb";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { InputTags } from "@/components/InputTags";
import { Switch } from "@/components/ui/switch";
import { SecretCopier } from "@/components/SecretCopier";
import { Title } from "@/components/Title";

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

  return (
    <>
      <Title title="API Keys" />
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

            <CreateAPIKeyButton />
          </div>
        </CardHeader>

        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>API Key ID</TableHead>
                <TableHead>Management API Access</TableHead>
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
                    <TableCell>
                      {apiKey?.hasManagementApiAccess ? "Yes" : "No"}
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

const APIKeyFormSchema = z.object({
  hasManagementAPIAccess: z.boolean(),
});

function CreateAPIKeyButton() {
  const { environmentId } = useParams();

  const { data: appOrganization } = useQuery(getAppOrganization, {});

  const form = useForm<z.infer<typeof APIKeyFormSchema>>({
    resolver: zodResolver(APIKeyFormSchema),
    defaultValues: {
      hasManagementAPIAccess: false,
    },
  });

  const [createFormOpen, setCreateFormOpen] = useState(false);
  const [apiKeyId, setApiKeyId] = useState("");
  const [apiKeySecret, setApiKeySecret] = useState("");
  const [showSecretAlertOpen, setShowSecretAlertOpen] = useState(false);
  const createAPIKeyMutation = useMutation(createAPIKey);
  const queryClient = useQueryClient();
  const handleSubmit = useCallback(
    async (values: z.infer<typeof APIKeyFormSchema>, e: any) => {
      e.preventDefault();
      const apiKey = await createAPIKeyMutation.mutateAsync({
        apiKey: {
          environmentId,
          hasManagementApiAccess: values.hasManagementAPIAccess,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey(listAPIKeys, { environmentId }),
      });

      setApiKeyId(apiKey.id);
      setApiKeySecret(apiKey.secretToken);
      setCreateFormOpen(false);
      setShowSecretAlertOpen(true);
    },
    [
      environmentId,
      createAPIKeyMutation,
      queryClient,
      setApiKeySecret,
      setShowSecretAlertOpen,
    ],
  );

  const getStripeCheckoutMutation = useMutation(getStripeCheckoutURL);
  const handleClickUpsell = async () => {
    const { url } = await getStripeCheckoutMutation.mutateAsync({});
    window.location.href = url;
  };

  return (
    <>
      <AlertDialog open={createFormOpen} onOpenChange={setCreateFormOpen}>
        <AlertDialogTrigger asChild>
          <Button variant="outline">
            <Plus className="mr-2 h-4 w-4" />
            Create API key
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(handleSubmit)}>
              <AlertDialogHeader>
                <AlertDialogTitle>Create API key</AlertDialogTitle>
              </AlertDialogHeader>

              <div className="my-4 space-y-4">
                <FormItem>
                  <FormLabel>SAML + SCIM API Access</FormLabel>
                  <FormControl className="block">
                    <Switch checked disabled />
                  </FormControl>
                  <FormDescription>
                    All API keys have SAML + SCIM API access. You can use these
                    keys to implement SAML or SCIM in your product.
                  </FormDescription>
                  <FormMessage />
                </FormItem>

                <FormField
                  control={form.control}
                  name="hasManagementAPIAccess"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Management API Access</FormLabel>
                      <FormControl className="block">
                        <Switch
                          name={field.name}
                          id={field.name}
                          checked={field.value}
                          onCheckedChange={field.onChange}
                          disabled={
                            appOrganization &&
                            !appOrganization.entitledManagementApi
                          }
                        />
                      </FormControl>
                      <FormDescription>
                        API Keys with management API access can programmatically
                        manage organizations, SAML connections, SCIM
                        directories, and customer self-serve setup URLs.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              {appOrganization && !appOrganization.entitledManagementApi && (
                <div className="my-4 p-4 bg-muted rounded text-sm">
                  <div className="font-semibold">
                    Management API is a Pro-tier feature
                  </div>
                  <Button
                    type="button"
                    className="mt-4"
                    onClick={handleClickUpsell}
                  >
                    Upgrade to Pro
                  </Button>
                </div>
              )}

              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <Button type="submit">Save</Button>
              </AlertDialogFooter>
            </form>
          </Form>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog
        open={showSecretAlertOpen}
        onOpenChange={setShowSecretAlertOpen}
      >
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

          <SecretCopier
            placeholder="ssoready_sk_•••••••••••••••••••••••••"
            secret={apiKeySecret}
          />

          <AlertDialogFooter>
            <AlertDialogAction asChild>
              <Link to={`/environments/${environmentId}/api-keys/${apiKeyId}`}>
                Done
              </Link>
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
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

          <SecretCopier
            placeholder="ssoready_oauth_client_secret_•••••••••••••••••••••••••"
            secret={samlOAuthClientSecret}
          />

          <AlertDialogFooter>
            <AlertDialogAction asChild>
              <Link
                to={`/environments/${environmentId}/saml-oauth-clients/${samlOAuthClientId}`}
              >
                Done
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
