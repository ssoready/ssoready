import { DocsLink } from "@/components/DocsLink";
import { InfoTooltip } from "@/components/InfoTooltip";
import { InputTags } from "@/components/InputTags";
import { SecretCopier } from "@/components/SecretCopier";
import { Title } from "@/components/Title";
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
import { Badge } from "@/components/ui/badge";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  appCreateAdminSetupURL,
  appCreateSAMLConnection,
  appCreateSCIMDirectory,
  appDeleteOrganization,
  appGetOrganization,
  appListSAMLConnections,
  appListSCIMDirectories,
  appUpdateOrganization,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Organization } from "@/gen/ssoready/v1/ssoready_pb";
import {
  createConnectQueryKey,
  useInfiniteQuery,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import React, { useCallback, useState } from "react";
import { useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router";
import { Link } from "react-router-dom";
import { toast } from "sonner";
import { z } from "zod";

export function ViewOrganizationPage() {
  const { environmentId, organizationId } = useParams();
  const { data: organization } = useQuery(appGetOrganization, {
    id: organizationId,
  });
  const {
    data: listSAMLConnectionResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    appListSAMLConnections,
    { organizationId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  const createSAMLConnectionMutation = useMutation(appCreateSAMLConnection);
  const navigate = useNavigate();
  const handleCreateSAMLConnection = useCallback(async () => {
    const samlConnection = await createSAMLConnectionMutation.mutateAsync({
      samlConnection: {
        organizationId,
        primary:
          listSAMLConnectionResponses?.pages.flatMap(
            (page) => page.samlConnections,
          ).length === 0,
      },
    });

    toast("SAML Connection has been created.");
    navigate(
      `/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnection.id}`,
    );
  }, [
    listSAMLConnectionResponses,
    environmentId,
    organizationId,
    createSAMLConnectionMutation,
    navigate,
  ]);

  return (
    <div className="flex flex-col gap-8">
      <Title title={organization?.displayName || "Organization"} />

      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbPage>
              {organization?.displayName || organizationId}
            </BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>
                  {organization?.displayName || "Organization"}
                </CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {organizationId}
                </span>
              </div>

              <CardDescription>
                An organization corresponds to a tenant in your application.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/organizations" />
              </CardDescription>
            </div>

            {organization && (
              <EditOrganizationAlertDialog organization={organization} />
            )}
          </div>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-5 gap-y-2">
            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              External ID
              <InfoTooltip>
                Optional. An ID from your database for this organization.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/organizations#organization-external-id" />
              </InfoTooltip>
            </div>
            <div className="text-sm col-span-3">
              {organization?.externalId || (
                <div className="text-sm text-muted-foreground">
                  Not configured
                </div>
              )}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              Domains
              <InfoTooltip>
                A whitelist of domains that customers from this organization
                use.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/organizations#domains" />
              </InfoTooltip>
            </div>
            <div className="text-sm col-span-3">
              {" "}
              <div className="flex gap-1 flex-wrap">
                {organization?.domains.map((domain, i) => (
                  <Badge key={i}>{domain}</Badge>
                ))}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <AdminSetupURLCard />

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <CardTitle>
                SAML Connections
                <DocsLink to="https://ssoready.com/docs/sso-ready-concepts/saml-connections" />
              </CardTitle>
              <CardDescription>
                SAML Connections within this organization.
              </CardDescription>
            </div>

            <Button variant="outline" onClick={handleCreateSAMLConnection}>
              <Plus className="mr-2 h-4 w-4" />
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
                        to={`/environments/${organization?.environmentId}/organizations/${organization?.id}/saml-connections/${samlConn.id}`}
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

      <OrganizationSCIMDirectoriesPage />
      <DangerZoneCard />
    </div>
  );
}

const AdminSetupURLFormSchema = z.object({
  canManageSAML: z.boolean(),
  canManageSCIM: z.boolean(),
});

function AdminSetupURLCard() {
  const { organizationId } = useParams();
  const createAdminSetupURLMutation = useMutation(appCreateAdminSetupURL);

  const [setupURL, setSetupURL] = useState("");
  const [showURLOpen, setShowURLOpen] = useState(false);

  const handleSubmit = async (
    values: z.infer<typeof AdminSetupURLFormSchema>,
    e: any,
  ) => {
    e.preventDefault();

    const { url } = await createAdminSetupURLMutation.mutateAsync({
      organizationId,
      canManageSaml: values.canManageSAML,
      canManageScim: values.canManageSCIM,
    });

    setSetupURL(url);
    setOpen(false);
    setShowURLOpen(true);
  };

  const [open, setOpen] = useState(false);

  const form = useForm<z.infer<typeof AdminSetupURLFormSchema>>({
    resolver: zodResolver(AdminSetupURLFormSchema),
    defaultValues: {
      canManageSAML: true,
      canManageSCIM: false,
    },
  });

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>
            Customer self-serve setup
            <DocsLink to="https://ssoready.com/docs/idp-configuration/enabling-self-service-configuration-for-your-customers" />
          </CardTitle>
          <CardDescription>
            You can invite your customer's IT admin to set up their SAML
            connection into your product. You can create for them a one-time-use
            link where they can create and modify this organization's SAML
            connections.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <AlertDialog open={open} onOpenChange={setOpen}>
            <AlertDialogTrigger asChild>
              <Button variant="outline">Create self-serve setup link</Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(handleSubmit)}>
                  <AlertDialogHeader>
                    <AlertDialogTitle>
                      Configure self-serve setup link
                    </AlertDialogTitle>
                  </AlertDialogHeader>

                  <div className="my-4 space-y-4">
                    <FormField
                      control={form.control}
                      name="canManageSAML"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Can manage SAML</FormLabel>
                          <FormControl className="block">
                            <Switch
                              name={field.name}
                              id={field.name}
                              checked={field.value}
                              onCheckedChange={field.onChange}
                            />
                          </FormControl>
                          <FormDescription>
                            Should your customer be able to configure SAML
                            connections?
                          </FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={form.control}
                      name="canManageSCIM"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Can manage SCIM</FormLabel>
                          <FormControl className="block">
                            <Switch
                              name={field.name}
                              id={field.name}
                              checked={field.value}
                              onCheckedChange={field.onChange}
                            />
                          </FormControl>
                          <FormDescription>
                            Should your customer be able to configure SCIM
                            directories?
                          </FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>

                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <Button type="submit">Create setup link</Button>
                  </AlertDialogFooter>
                </form>
              </Form>
            </AlertDialogContent>
          </AlertDialog>
        </CardContent>
      </Card>

      <AlertDialog open={showURLOpen} onOpenChange={setShowURLOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Customer self-serve setup url created
            </AlertDialogTitle>
            <AlertDialogDescription>
              Your customer self-serve setup URL has been created.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <SecretCopier
            placeholder={
              setupURL &&
              `https://${new URL(setupURL).hostname}/setup?one-time-token=...`
            }
            secret={setupURL}
          />

          <p className="text-sm">
            Copy this setup URL and give it to your customer. It'll take them to
            a page where they can configure their SAML/SCIM settings.
          </p>

          <p className="text-sm">
            Setup URLs are one-time-use only. If you visit this URL yourself,
            you'll need to generate a new one to give to your customer.
          </p>

          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction>Done</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

const FormSchema = z.object({
  displayName: z.string(),
  externalId: z.string(),
  domains: z.array(z.string()).min(1, {
    message: "At least one domain is required.",
  }),
});

function EditOrganizationAlertDialog({
  organization,
}: {
  organization: Organization;
}) {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      displayName: organization.displayName,
      externalId: organization.externalId,
      domains: organization.domains,
    },
  });

  const [open, setOpen] = useState(false);
  const updateOrganizationMutation = useMutation(appUpdateOrganization);
  const queryClient = useQueryClient();
  const handleSubmit = useCallback(
    async (values: z.infer<typeof FormSchema>, e: any) => {
      e.preventDefault();
      await updateOrganizationMutation.mutateAsync({
        organization: {
          id: organization.id,
          displayName: values.displayName,
          externalId: values.externalId,
          domains: values.domains,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey(appGetOrganization, {
          id: organization.id,
        }),
      });

      setOpen(false);
    },
    [setOpen, organization, updateOrganizationMutation, queryClient],
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
              <AlertDialogTitle>Edit organization</AlertDialogTitle>
            </AlertDialogHeader>

            <div className="my-4 space-y-4">
              <FormField
                control={form.control}
                name="displayName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>DisplayName</FormLabel>
                    <FormControl>
                      <Input placeholder="AcmeCorp" {...field} />
                    </FormControl>
                    <FormDescription>
                      An optional human-friendly name for the organization.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

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

function OrganizationSCIMDirectoriesPage() {
  const { environmentId, organizationId } = useParams();
  const { data: organization } = useQuery(appGetOrganization, {
    id: organizationId,
  });
  const {
    data: listSCIMDirectoriesPages,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    appListSCIMDirectories,
    { organizationId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  const createSCIMDirectoryMutation = useMutation(appCreateSCIMDirectory);
  const navigate = useNavigate();
  const handleCreateSCIMDirectory = useCallback(async () => {
    const scimDirectory = await createSCIMDirectoryMutation.mutateAsync({
      scimDirectory: {
        organizationId,
        primary:
          listSCIMDirectoriesPages?.pages.flatMap(
            (page) => page.scimDirectories,
          ).length === 0,
      },
    });

    toast("SCIM Directory has been created.");
    navigate(
      `/environments/${environmentId}/organizations/${organizationId}/scim-directories/${scimDirectory.id}`,
    );
  }, [
    // listSAMLConnectionResponses,
    environmentId,
    organizationId,
    createSCIMDirectoryMutation,
    navigate,
  ]);

  return (
    <Card>
      <CardHeader>
        <div className="flex justify-between items-center">
          <div className="flex flex-col space-y-1.5">
            <CardTitle>
              SCIM Directories
              <DocsLink to="https://ssoready.com/docs/ssoready-concepts/scim-directories" />
            </CardTitle>
            <CardDescription>
              SCIM Directories within this organization.
            </CardDescription>
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
                      to={`/environments/${organization?.environmentId}/organizations/${organization?.id}/scim-directories/${scimDirectory.id}`}
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
  );
}

function DangerZoneCard() {
  const { environmentId, organizationId } = useParams();
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);

  const handleDelete = () => {
    setConfirmDeleteOpen(true);
  };

  const deleteOrganizationMutation = useMutation(appDeleteOrganization);
  const navigate = useNavigate();
  const handleConfirmDelete = async () => {
    await deleteOrganizationMutation.mutateAsync({
      organizationId,
    });

    toast.success("Organization deleted");
    navigate(`/environments/${environmentId}`);
  };

  return (
    <>
      <AlertDialog open={confirmDeleteOpen} onOpenChange={setConfirmDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete organization?</AlertDialogTitle>
            <AlertDialogDescription>
              Deleting an organization cannot be undone. All SAML connections
              and SCIM directories inside the organization will be permanently
              deleted as well.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <Button variant="destructive" onClick={handleConfirmDelete}>
              Permanently Delete Organization
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Card className="border-destructive">
        <CardHeader>
          <CardTitle>Danger Zone</CardTitle>
        </CardHeader>

        <CardContent>
          <div className="flex justify-between items-center">
            <div>
              <div className="text-sm font-semibold">Delete Organization</div>
              <p className="text-sm">
                Delete this organization. This cannot be undone.
              </p>
            </div>

            <Button variant="destructive" onClick={handleDelete}>
              Delete organization
            </Button>
          </div>
        </CardContent>
      </Card>
    </>
  );
}
