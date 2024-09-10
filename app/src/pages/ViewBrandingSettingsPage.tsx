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
  CardFooter,
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
import { Skeleton } from "@/components/ui/skeleton";
import { Link } from "react-router-dom";
import { Title } from "@/components/Title";
import { DocsLink } from "@/components/DocsLink";

export function ViewBrandingSettingsPage() {
  const { environmentId } = useParams();
  const { data: environmentAdminSettings } = useQuery(appGetAdminSettings, {
    environmentId,
  });

  return (
    <div className="flex flex-col gap-y-8">
      <Title title="Branding" />

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
            <DocsLink to="https://ssoready.com/docs/ssoready-concepts/environments#custom-branding" />
          </CardDescription>
        </CardHeader>

        <CardContent>
          <div className="grid grid-cols-5 gap-y-2 items-center">
            <div className="text-sm col-span-2 text-muted-foreground">
              Application Name
            </div>
            <div className="text-sm col-span-3">
              {environmentAdminSettings?.adminApplicationName || (
                <div className="text-sm text-muted-foreground">
                  Not configured
                </div>
              )}
            </div>

            <div className="text-sm col-span-2 text-muted-foreground">
              Return URL
            </div>
            <div className="text-sm col-span-3">
              {environmentAdminSettings?.adminReturnUrl || (
                <div className="text-sm text-muted-foreground">
                  Not configured
                </div>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Preview</CardTitle>
          <CardDescription>
            Here's a preview of what your customers will see when they're
            setting up SAML or SCIM settings.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <div className="border bg-muted p-8">
            <Card>
              <nav className="border-b border-gray-200 bg-white rounded-t-lg">
                <div className="mx-4 flex h-16 justify-between">
                  <div className="flex">
                    <span className="flex flex-shrink-0 items-center text-sm">
                      {environmentAdminSettings?.adminLogoUrl && (
                        <img
                          className="h-8 w-8 mr-4"
                          src={environmentAdminSettings.adminLogoUrl}
                          alt=""
                        />
                      )}

                      {environmentAdminSettings?.adminApplicationName
                        ? environmentAdminSettings.adminApplicationName
                        : "Settings Panel"}
                    </span>

                    <div className="ml-6 flex space-x-8">
                      <span className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center border-b-2 px-1 pt-1 text-sm font-medium">
                        SAML Settings
                      </span>
                      <span className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center border-b-2 px-1 pt-1 text-sm font-medium">
                        SCIM Settings
                      </span>
                    </div>
                  </div>

                  {environmentAdminSettings?.adminReturnUrl && (
                    <Link
                      to={environmentAdminSettings.adminReturnUrl}
                      className="flex items-center text-sm ml-auto"
                    >
                      {environmentAdminSettings?.adminApplicationName
                        ? `Back to ${environmentAdminSettings.adminApplicationName}`
                        : "Exit"}
                    </Link>
                  )}
                </div>
              </nav>

              <CardHeader>
                <Skeleton className="animate-none h-6 w-32 rounded" />
              </CardHeader>
              <CardContent>
                <Skeleton className="animate-none h-4 w-full mb-2" />
                <Skeleton className="animate-none h-4 w-full mb-2" />
                <Skeleton className="animate-none h-4 w-full mb-2" />
              </CardContent>
              <CardFooter>
                <Skeleton className="animate-none h-8 w-24 rounded" />
              </CardFooter>
            </Card>
          </div>
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

    if (logo.current!.files![0]) {
      const { uploadUrl } = await updateAdminSettingsLogoMutation.mutateAsync({
        environmentId,
      });

      await fetch(uploadUrl, {
        method: "PUT",
        body: logo.current!.files![0],
      });
    }

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
                    <FormDescription>
                      A display name for your application / company.
                    </FormDescription>
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
                      When your customers are done setting things up, this is
                      where an "Exit this setup UI" link should take them. Make
                      this point at your web application.
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
                    <FormDescription>Your company's logo.</FormDescription>
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
