import React, { useCallback, useState } from "react";
import { useParams } from "react-router";
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  createOrganization,
  getEnvironment,
  listOrganizations,
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
import { Plus } from "lucide-react";
import { InputTags } from "@/components/InputTags";

export function ViewEnvironmentPage() {
  const { environmentId } = useParams();
  const { data: environment } = useQuery(getEnvironment, {
    id: environmentId,
  });
  const { data: listOrgsRes } = useQuery(listOrganizations, {
    environmentId,
  });

  return (
    <div className="grid gap-8">
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
          <div className="grid grid-cols-4">
            <div className="text-sm col-span-1 text-muted-foreground">
              Redirect URL
            </div>
            <div className="text-sm col-span-3">{environment?.redirectUrl}</div>
          </div>
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
              {listOrgsRes?.organizations?.map((org) => (
                <TableRow key={org.id}>
                  <TableCell>
                    <Link
                      to={`/environments/${environmentId}/organizations/${org?.id}`}
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
  domains: z.array(z.string()),
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
  const handleSubmit = useCallback(
    async (values: z.infer<typeof OrgFormSchema>, e: any) => {
      e.preventDefault();
      await createOrganizationMutation.mutateAsync({
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
    },
    [setOpen, environment, createOrganizationMutation, queryClient],
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
