import React, { useCallback, useRef, useState } from "react";
import { useParams } from "react-router";
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  appGetAdminSettings,
  appUpdateAdminSettings,
  appUpdateAdminSettingsLogo,
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
import {
  AppGetAdminSettingsResponse,
  Environment,
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
  const { data: environmentAdminSettings } = useQuery(appGetAdminSettings, {
    environmentId,
  });

  return (
    <div className="flex flex-col gap-y-8">
      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <CardTitle>Self-Serve Setup UI Branding</CardTitle>

            <div>
              {environmentAdminSettings && (
                <EditSettingsAlertDialog
                  environmentAdminSettings={environmentAdminSettings}
                />
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
              {environmentAdminSettings?.adminApplicationName}
            </div>

            <div className="text-sm col-span-1 text-muted-foreground">
              Return URL
            </div>
            <div className="text-sm col-span-3">
              {environmentAdminSettings?.adminReturnUrl}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Preview</CardTitle>
        </CardHeader>

        <CardContent>
          {environmentAdminSettings?.adminLogoUrl && (
            <img
              className="h-8 w-8"
              src={environmentAdminSettings.adminLogoUrl}
              alt=""
            />
          )}
        </CardContent>
      </Card>
    </div>
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
  logo: z.string().optional(),
});

function EditSettingsAlertDialog({
  environmentAdminSettings,
}: {
  environmentAdminSettings: AppGetAdminSettingsResponse;
}) {
  const { environmentId } = useParams();
  const logo = useRef<HTMLInputElement>(null);
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      applicationName: environmentAdminSettings.adminApplicationName,
      returnUrl: environmentAdminSettings.adminReturnUrl,
    },
  });

  const [open, setOpen] = useState(false);
  const updateAdminSettingsMutation = useMutation(appUpdateAdminSettings);
  const updateAdminSettingsLogoMutation = useMutation(
    appUpdateAdminSettingsLogo,
  );
  const queryClient = useQueryClient();
  const handleSubmit = async (values: z.infer<typeof formSchema>, e: any) => {
    e.preventDefault();
    await updateAdminSettingsMutation.mutateAsync({
      environmentId,
      adminApplicationName: values.applicationName,
      adminReturnUrl: values.returnUrl,
    });

    const { uploadUrl } = await updateAdminSettingsLogoMutation.mutateAsync({
      environmentId,
    });

    await fetch(uploadUrl, {
      method: "PUT",
      body: logo.current!.files![0],
    });

    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey(appGetAdminSettings, {
        environmentId,
      }),
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

              <FormField
                control={form.control}
                name="logo"
                render={() => (
                  <FormItem>
                    <FormLabel>Logo</FormLabel>
                    <FormControl>
                      <Input type="file" accept="image/*" ref={logo} />
                    </FormControl>
                    <FormDescription>asdf</FormDescription>
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
