import React, { useCallback, useState } from "react";
import { useParams } from "react-router";
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  appUpdateAdminSettings,
  getEnvironment,
  updateEnvironment,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { z } from "zod";
import { Environment } from "@/gen/ssoready/v1/ssoready_pb";
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
import { Button } from "@/components/ui/button";
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

export function ViewBrandingSettingsPage() {
  const { environmentId } = useParams();
  const { data: environment } = useQuery(getEnvironment, {
    id: environmentId,
  });

  return (
    <Card>
      <CardHeader>
        <div className="flex justify-between items-center">
          <CardTitle>Self-Serve Setup UI Branding</CardTitle>

          <div>
            {environment && (
              <EditSettingsAlertDialog environment={environment} />
            )}
          </div>
        </div>
        <CardDescription>
          SSOReady can host a UI on your behalf that lets your customers
          configure their SAML and SCIM settings on their own. Here, you can
          make that UI look on-brand for your company.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <div className="grid grid-cols-4 gap-y-2 items-center">
          <div className="text-sm col-span-1 text-muted-foreground">
            Application Name
          </div>
          <div className="text-sm col-span-3">
            {environment?.adminApplicationName}
          </div>

          <div className="text-sm col-span-1 text-muted-foreground">
            Return URL
          </div>
          <div className="text-sm col-span-3">
            {environment?.adminReturnUrl}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

const formSchema = z.object({
  applicationName: z
    .string()
    .min(1, { message: "Application name is required." }),
  returnUrl: z
    .string()
    .min(1, { message: "Return URL is required." })
    .url({ message: "Return URL must be a valid URL." }),
});

function EditSettingsAlertDialog({
  environment,
}: {
  environment: Environment;
}) {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      applicationName: environment.adminApplicationName,
      returnUrl: environment.adminReturnUrl,
    },
  });

  const [open, setOpen] = useState(false);
  const updateAdminSettingsMutation = useMutation(appUpdateAdminSettings);
  const queryClient = useQueryClient();
  const handleSubmit = async (values: z.infer<typeof formSchema>, e: any) => {
    e.preventDefault();
    await updateAdminSettingsMutation.mutateAsync({
      environmentId: environment.id,
      adminApplicationName: values.applicationName,
      adminReturnUrl: values.returnUrl,
    });

    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey(getEnvironment, { id: environment.id }),
    });

    setOpen(false);
  };

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        <Button variant="outline">Edit</Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)}>
            <AlertDialogHeader>
              <AlertDialogTitle>
                Edit environment branding settings
              </AlertDialogTitle>
            </AlertDialogHeader>

            <div className="my-4 space-y-4">
              <FormField
                control={form.control}
                name="applicationName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Application name</FormLabel>
                    <FormControl>
                      <Input placeholder="Your Company Name" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="returnUrl"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Return URL</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="https://app.yourcompany.com"
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
