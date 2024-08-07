import React, { useCallback, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router";
import {
  createConnectQueryKey,
  useInfiniteQuery,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  createAdminSetupURL,
  createOrganization,
  createSAMLConnection,
  getEnvironment,
  getOrganization,
  getSAMLConnection,
  listOrganizations,
  listSAMLConnections,
  listSAMLFlows,
  updateEnvironment,
  updateOrganization,
  updateSAMLConnection,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
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
import { Plus, PlusCircle } from "lucide-react";
import { z } from "zod";
import {
  Environment,
  Organization,
  SAMLConnection,
} from "@/gen/ssoready/v1/ssoready_pb";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
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
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { DocsLink } from "@/components/DocsLink";

export function ViewOrganizationPage() {
  const { environmentId, organizationId } = useParams();
  const { data: organization } = useQuery(getOrganization, {
    id: organizationId,
  });
  const {
    data: listSAMLConnectionResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    listSAMLConnections,
    { organizationId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  const createSAMLConnectionMutation = useMutation(createSAMLConnection);
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

  const createAdminSetupURLMutation = useMutation(createAdminSetupURL);
  const handleCreateAdminSetupURL = async () => {
    const { url } = await createAdminSetupURLMutation.mutateAsync({
      organizationId,
    });

    await navigator.clipboard.writeText(url);
    toast("Setup link copied to clipboard");
  };

  return (
    <div className="flex flex-col gap-8">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbPage>{organizationId}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>Organization</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {organizationId}
                </span>
              </div>

              <CardDescription>
                An organization corresponds to a tenant in your application.
              </CardDescription>
            </div>

            {organization && (
              <EditOrganizationAlertDialog organization={organization} />
            )}
          </div>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-4 gap-y-2">
            <div className="text-sm col-span-1 text-muted-foreground">
              External ID
            </div>
            <div className="text-sm col-span-3">{organization?.externalId}</div>

            <div className="text-sm col-span-1 text-muted-foreground">
              Domains
            </div>
            <div className="text-sm col-span-3">
              {" "}
              <div className="flex gap-1">
                {organization?.domains.map((domain, i) => (
                  <Badge key={i}>{domain}</Badge>
                ))}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex gap-x-4 items-center">
            <span>Customer self-serve setup</span>
            <Badge variant="secondary">Beta</Badge>
          </CardTitle>
          <CardDescription>
            You can invite your customer's IT admin to set up their SAML
            connection into your product. You can create for them a one-time-use
            link where they can create and modify this organization's SAML
            connections.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <Button variant="outline" onClick={handleCreateAdminSetupURL}>
            Copy setup link
          </Button>
        </CardContent>
      </Card>

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
    </div>
  );
}

const FormSchema = z.object({
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
      externalId: organization.externalId,
      domains: organization.domains,
    },
  });

  const [open, setOpen] = useState(false);
  const updateOrganizationMutation = useMutation(updateOrganization);
  const queryClient = useQueryClient();
  const handleSubmit = useCallback(
    async (values: z.infer<typeof FormSchema>, e: any) => {
      e.preventDefault();
      await updateOrganizationMutation.mutateAsync({
        organization: {
          id: organization.id,
          externalId: values.externalId,
          domains: values.domains,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey(getOrganization, {
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
