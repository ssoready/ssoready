import React, { useCallback, useState } from "react";
import { useNavigate, useParams } from "react-router";
import {
  createConnectQueryKey,
  useInfiniteQuery,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  createAPIKey,
  createOrganization,
  getEnvironment,
  listAPIKeys,
  listOrganizations,
  listSAMLConnections,
  updateEnvironment,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
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
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Environment } from "@/gen/ssoready/v1/ssoready_pb";
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
import { useQueryClient } from "@tanstack/react-query";
import { CircleAlert, Plus } from "lucide-react";
import { InputTags } from "@/components/InputTags";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { toast } from "sonner";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

export function ViewEnvironmentPage() {
  const { environmentId } = useParams();
  const { data: environment } = useQuery(getEnvironment, {
    id: environmentId,
  });
  const { data: listAPIKeysRes } = useQuery(listAPIKeys, {
    environmentId,
  });
  const {
    data: listOrganizationsResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    listOrganizations,
    { environmentId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  const [apiKeySecret, setApiKeySecret] = useState("");
  const [apiKeyAlertOpen, setApiKeyAlertOpen] = useState(false);
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

    setApiKeySecret(apiKey.secretToken);
    setApiKeyAlertOpen(true);
  }, [
    environmentId,
    createAPIKeyMutation,
    queryClient,
    setApiKeySecret,
    setApiKeyAlertOpen,
  ]);

  return (
    <div className="flex flex-col gap-y-8">
      <AlertDialog open={apiKeyAlertOpen} onOpenChange={setApiKeyAlertOpen}>
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
            <AlertDialogCancel>Close</AlertDialogCancel>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>{environment?.displayName}</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {environmentId}
                </span>
              </div>

              <CardDescription>
                An environment corresponds to a deployment environment in your
                product, e.g. "Staging" or "Production".
              </CardDescription>
            </div>

            {environment && (
              <EditEnvironmentAlertDialog environment={environment} />
            )}
          </div>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-4 gap-y-2">
            <div className="text-sm col-span-1 text-muted-foreground">
              Redirect URL
            </div>
            <div className="text-sm col-span-3">{environment?.redirectUrl}</div>
            <div className="text-sm col-span-1 text-muted-foreground">
              Auth URL
            </div>
            <div className="text-sm col-span-3">{environment?.authUrl}</div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <CardTitle>API Keys</CardTitle>

              <CardDescription>API keys for this environment.</CardDescription>
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
                <TableHead>API Key ID</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {listAPIKeysRes?.apiKeys?.map((apiKey) => (
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
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <CardTitle>Organizations</CardTitle>

              <CardDescription>
                Organizations within this environment.
              </CardDescription>
            </div>

            {environment && (
              <CreateOrganizationAlertDialog environment={environment} />
            )}
          </div>
        </CardHeader>

        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Organization ID</TableHead>
                <TableHead>External ID</TableHead>
                <TableHead>Domains</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {listOrganizationsResponses?.pages
                .flatMap((page) => page.organizations)
                .map((org) => (
                  <TableRow key={org.id}>
                    <TableCell>
                      <Link
                        to={`/environments/${environmentId}/organizations/${org?.id}`}
                        className="underline underline-offset-4 decoration-muted-foreground"
                      >
                        {org.id}
                      </Link>
                    </TableCell>
                    <TableCell>{org.externalId}</TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        {org.domains.map((domain, i) => (
                          <Badge key={i}>{domain}</Badge>
                        ))}
                      </div>
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
    </div>
  );
}

const FormSchema = z.object({
  displayName: z.string().min(1, {
    message: "Display name is required.",
  }),
  redirectUrl: z.string().url({
    message: "Redirect URL must be a valid URL.",
  }),
  authUrl: z.string().url({
    message: "Auth URL must be a valid URL.",
  }),
});

function EditEnvironmentAlertDialog({
  environment,
}: {
  environment: Environment;
}) {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      displayName: environment.displayName,
      redirectUrl: environment.redirectUrl,
      authUrl: environment.authUrl,
    },
  });

  const [open, setOpen] = useState(false);
  const updateEnvironmentMutation = useMutation(updateEnvironment);
  const queryClient = useQueryClient();
  const handleSubmit = useCallback(
    async (values: z.infer<typeof FormSchema>, e: any) => {
      e.preventDefault();
      await updateEnvironmentMutation.mutateAsync({
        environment: {
          id: environment.id,
          displayName: values.displayName,
          redirectUrl: values.redirectUrl,
          authUrl: values.authUrl,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey(getEnvironment, { id: environment.id }),
      });

      setOpen(false);
    },
    [setOpen, environment, updateEnvironmentMutation, queryClient],
  );

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        <Button variant="outline">Edit</Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)}>
            <AlertDialogHeader>
              <AlertDialogTitle>Edit environment</AlertDialogTitle>
            </AlertDialogHeader>

            <div className="my-4 space-y-4">
              <FormField
                control={form.control}
                name="displayName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Display name</FormLabel>
                    <FormControl>
                      <Input placeholder="Production" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="redirectUrl"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Redirect URL</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="http://localhost:8080/ssoready/callback"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      After a SAML login, your users get redirected to this
                      address. You usually want to point this at an
                      SSOReady-specific page on your web application.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="authUrl"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Auth URL</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormDescription>
                      The base of the URL that your customer's IT admins put
                      into their Identity Provider. This should either be
                      https://auth.ssoready.com or a custom domain that CNAMEs
                      there.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <Button type="submit">Save</Button>
            </AlertDialogFooter>
          </form>
        </Form>
      </AlertDialogContent>
    </AlertDialog>
  );
}

const OrgFormSchema = z.object({
  externalId: z.string(),
  domains: z.array(z.string()).min(1, {
    message: "At least one domain is required.",
  }),
});

function CreateOrganizationAlertDialog({
  environment,
}: {
  environment: Environment;
}) {
  const form = useForm<z.infer<typeof OrgFormSchema>>({
    resolver: zodResolver(OrgFormSchema),
    defaultValues: {
      externalId: "",
      domains: [],
    },
  });

  const [open, setOpen] = useState(false);
  const createOrganizationMutation = useMutation(createOrganization);
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const handleSubmit = useCallback(
    async (values: z.infer<typeof OrgFormSchema>, e: any) => {
      e.preventDefault();
      const organization = await createOrganizationMutation.mutateAsync({
        organization: {
          environmentId: environment.id,
          externalId: values.externalId,
          domains: values.domains,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey(listOrganizations, {
          environmentId: environment.id,
        }),
      });

      setOpen(false);
      toast("Organization has been created.");
      navigate(
        `/environments/${environment.id}/organizations/${organization.id}`,
      );
    },
    [setOpen, environment, createOrganizationMutation, queryClient, navigate],
  );

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        <Button variant="outline">
          <Plus className="mr-2 h-4 w-4" />
          Create organization
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)}>
            <AlertDialogHeader>
              <AlertDialogTitle>Edit environment</AlertDialogTitle>
            </AlertDialogHeader>

            <div className="my-4 space-y-4">
              <FormField
                control={form.control}
                name="externalId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>External ID</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. 42, 507f191e810c19729de860ea, ..."
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      An optional unique identifier for this organization. This
                      is returned in the SAML Redeem endpoint. Use this to more
                      easily tie an SSOReady organization to its counterpart in
                      your application.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="domains"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Domains</FormLabel>
                    <FormControl>
                      <InputTags {...field} />
                    </FormControl>
                    <FormDescription>
                      SSOReady will only allow SAML logins from users whose
                      email are in this list of domains.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <Button type="submit">Save</Button>
            </AlertDialogFooter>
          </form>
        </Form>
      </AlertDialogContent>
    </AlertDialog>
  );
}
