import React, { useCallback, useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useParams } from "react-router";
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  checkEnvironmentCustomDomainSettingsCertificates,
  getAppOrganization,
  getEnvironmentCustomDomainSettings,
  getStripeCheckoutURL,
  updateEnvironmentCustomDomainSettings,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Badge } from "@/components/ui/badge";
import {
  ArrowRightIcon,
  CopyIcon,
  GlobeIcon,
  LoaderCircleIcon,
  PlusIcon,
} from "lucide-react";
import { offset, useFloating, useTransitionStyles } from "@floating-ui/react";
import { Button } from "@/components/ui/button";
import { z } from "zod";
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
import { GetEnvironmentCustomDomainSettingsResponse } from "@/gen/ssoready/v1/ssoready_pb";
import { Title } from "@/components/Title";

export function ViewCustomDomainsPage() {
  const { environmentId } = useParams();

  const { data: appOrganization } = useQuery(getAppOrganization, {});

  const { data: customDomainsSettings } = useQuery(
    getEnvironmentCustomDomainSettings,
    {
      environmentId,
    },
  );

  const checkCertificateMutation = useMutation(
    checkEnvironmentCustomDomainSettingsCertificates,
  );
  const queryClient = useQueryClient();
  const handleCheckCertificate = async () => {
    await checkCertificateMutation.mutateAsync({
      environmentId,
    });

    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey(getEnvironmentCustomDomainSettings, {
        environmentId,
      }),
    });
  };

  const getStripeCheckoutMutation = useMutation(getStripeCheckoutURL);
  const handleClickUpsell = async () => {
    const { url } = await getStripeCheckoutMutation.mutateAsync({});
    window.location.href = url;
  };

  return (
    <div className="flex flex-col gap-y-8">
      <Title title="Custom Domains" />

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <CardTitle>
              Custom Domain for{" "}
              <code className="text-xl bg-gray-100 py-1 px-2 rounded-sm">
                auth.ssoready.com
              </code>
            </CardTitle>
            {appOrganization?.entitledCustomDomains &&
              customDomainsSettings && (
                <EditCustomAuthDomainAlertModal
                  customDomainSettings={customDomainsSettings}
                />
              )}
          </div>
          <CardDescription>
            SAML and SCIM both require that you set up a server that your
            customer's Identity Provider talks to. SSOReady runs that server for
            you. By default, it's hosted at auth.ssoready.com. With a custom
            domain, you can run it on any domain you control.
          </CardDescription>
        </CardHeader>

        <CardContent>
          {appOrganization?.entitledCustomDomains && customDomainsSettings && (
            <>
              {customDomainsSettings.customAuthDomain ? (
                <>
                  <div>
                    <div className="text-sm font-semibold mb-2">
                      Custom auth.ssoready.com domain
                    </div>
                    <div className="text-sm mb-4">
                      {customDomainsSettings.customAuthDomain}
                    </div>

                    <div className="text-sm font-semibold mb-2">
                      DNS record status
                    </div>
                    {customDomainsSettings.customAuthDomainConfigured ? (
                      <Badge
                        variant="outline"
                        className="inline-flex items-center gap-x-2"
                      >
                        <span className="relative flex h-3 w-3">
                          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                          <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                        </span>
                        <span>Live</span>
                      </Badge>
                    ) : (
                      <div>
                        <div>
                          <Badge variant="outline">
                            Waiting for your DNS changes
                          </Badge>
                        </div>

                        <Button
                          onClick={handleCheckCertificate}
                          variant="outline"
                          className="mt-2 inline-flex gap-x-2"
                        >
                          {checkCertificateMutation.isPending && (
                            <LoaderCircleIcon className="h-4 w-4 animate-spin" />
                          )}
                          Check now
                        </Button>
                      </div>
                    )}
                  </div>

                  {customDomainsSettings.customAuthDomainConfigured ? (
                    <p className="mt-4 text-sm text-muted-foreground">
                      Your custom domain is active, and all new SAML connections
                      and SCIM directories will use it.
                    </p>
                  ) : (
                    <>
                      <div className="mt-4 grid grid-cols-[1fr_auto_1fr] gap-x-4">
                        <div className="text-sm font-semibold mb-2">
                          DNS record name
                        </div>
                        <div></div>
                        <div className="text-sm font-semibold mb-2">
                          DNS record value
                        </div>
                        <div className="flex items-center text-sm border rounded">
                          <div className="py-2 px-4 bg-muted">CNAME</div>
                          <div className="flex-1 py-2 px-4">
                            {customDomainsSettings?.customAuthDomain}
                          </div>
                          <div className="px-2">
                            {customDomainsSettings && (
                              <CopyButton
                                value={customDomainsSettings.customAuthDomain}
                              />
                            )}
                          </div>
                        </div>
                        <div className="flex items-center">
                          <ArrowRightIcon className="h-5 w-5 text-muted-foreground" />
                        </div>
                        <div className="flex items-center text-sm border rounded">
                          <div className="flex-1 py-2 px-4">
                            {customDomainsSettings?.customAuthDomainCnameValue}
                          </div>
                          <div className="px-2">
                            {customDomainsSettings && (
                              <CopyButton
                                value={
                                  customDomainsSettings.customAuthDomainCnameValue
                                }
                              />
                            )}
                          </div>
                        </div>
                      </div>

                      <p className="mt-4 text-sm text-muted-foreground">
                        To use a custom domain, you first need to configure a
                        CNAME. Custom domains must use HTTPS for security, and
                        we can't provision a TLS certificate for your custom
                        domain until you've updated your DNS.
                      </p>
                    </>
                  )}
                </>
              ) : (
                <div className="flex flex-col items-center">
                  <GlobeIcon className="h-10 w-10 text-muted-foreground" />
                  <div className="text-sm font-semibold">No custom domain</div>
                  <div className="text-sm text-muted-foreground">
                    You do not have a custom domain configured. Click "Edit"
                    above to configure one.
                  </div>
                </div>
              )}
            </>
          )}

          {appOrganization && !appOrganization.entitledCustomDomains && (
            <div className="my-4 p-4 bg-muted rounded text-sm">
              <div className="font-semibold">
                Custom Domains is a Pro-tier feature
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
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <CardTitle>
              Custom Domain for{" "}
              <code className="text-xl bg-gray-100 py-1 px-2 rounded-sm">
                admin.ssoready.com
              </code>
            </CardTitle>
            {appOrganization?.entitledCustomDomains &&
              customDomainsSettings && (
                <EditCustomAdminDomainAlertModal
                  customDomainSettings={customDomainsSettings}
                />
              )}
          </div>
          <CardDescription>
            SSOReady provides an app that can guide your customer through the
            set-up process for SAML Connections and SCIM Directories. By
            default, it's hosted at admin.ssoready.com. With a custom domain,
            you can run it on any domain you control.
          </CardDescription>
        </CardHeader>

        <CardContent>
          {appOrganization?.entitledCustomDomains && customDomainsSettings && (
            <>
              {customDomainsSettings.customAdminDomain ? (
                <>
                  <div>
                    <div className="text-sm font-semibold mb-2">
                      Custom admin.ssoready.com domain
                    </div>
                    <div className="text-sm mb-4">
                      {customDomainsSettings.customAdminDomain}
                    </div>

                    <div className="text-sm font-semibold mb-2">
                      DNS record status
                    </div>
                    {customDomainsSettings.customAdminDomainConfigured ? (
                      <Badge
                        variant="outline"
                        className="inline-flex items-center gap-x-2"
                      >
                        <span className="relative flex h-3 w-3">
                          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                          <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                        </span>
                        <span>Live</span>
                      </Badge>
                    ) : (
                      <div>
                        <div>
                          <Badge variant="outline">
                            Waiting for your DNS changes
                          </Badge>
                        </div>

                        <Button
                          onClick={handleCheckCertificate}
                          variant="outline"
                          className="mt-2 inline-flex gap-x-2"
                        >
                          {checkCertificateMutation.isPending && (
                            <LoaderCircleIcon className="h-4 w-4 animate-spin" />
                          )}
                          Check now
                        </Button>
                      </div>
                    )}
                  </div>

                  {customDomainsSettings.customAdminDomainConfigured ? (
                    <p className="mt-4 text-sm text-muted-foreground">
                      Your custom domain for admin.ssoready.com is active. All
                      new self-serve setup links use your custom domain.
                    </p>
                  ) : (
                    <>
                      <div className="mt-4 grid grid-cols-[1fr_auto_1fr] gap-x-4">
                        <div className="text-sm font-semibold mb-2">
                          DNS record name
                        </div>
                        <div></div>
                        <div className="text-sm font-semibold mb-2">
                          DNS record value
                        </div>
                        <div className="flex items-center text-sm border rounded">
                          <div className="py-2 px-4 bg-muted">CNAME</div>
                          <div className="flex-1 py-2 px-4">
                            {customDomainsSettings?.customAdminDomain}
                          </div>
                          <div className="px-2">
                            {customDomainsSettings && (
                              <CopyButton
                                value={customDomainsSettings.customAdminDomain}
                              />
                            )}
                          </div>
                        </div>
                        <div className="flex items-center">
                          <ArrowRightIcon className="h-5 w-5 text-muted-foreground" />
                        </div>
                        <div className="flex items-center text-sm border rounded">
                          <div className="flex-1 py-2 px-4">
                            {customDomainsSettings?.customAdminDomainCnameValue}
                          </div>
                          <div className="px-2">
                            {customDomainsSettings && (
                              <CopyButton
                                value={
                                  customDomainsSettings.customAdminDomainCnameValue
                                }
                              />
                            )}
                          </div>
                        </div>
                      </div>

                      <p className="mt-4 text-sm text-muted-foreground">
                        To use a custom domain, you first need to configure a
                        CNAME. Custom domains must use HTTPS for security, and
                        we can't provision a TLS certificate for your custom
                        domain until you've updated your DNS.
                      </p>
                    </>
                  )}
                </>
              ) : (
                <div className="flex flex-col items-center">
                  <GlobeIcon className="h-10 w-10 text-muted-foreground" />
                  <div className="text-sm font-semibold">No custom domain</div>
                  <div className="text-sm text-muted-foreground">
                    You do not have a custom domain configured. Click "Edit"
                    above to configure one.
                  </div>
                </div>
              )}
            </>
          )}

          {appOrganization && !appOrganization.entitledCustomDomains && (
            <div className="my-4 p-4 bg-muted rounded text-sm">
              <div className="font-semibold">
                Custom Domains is a Pro-tier feature
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
        </CardContent>
      </Card>
    </div>
  );
}

function CopyButton({ value }: { value: string }) {
  const [open, setOpen] = useState(false);
  const { refs, floatingStyles, context } = useFloating({
    open,
    onOpenChange: setOpen,
    placement: "top",
    middleware: [offset(5)],
  });
  const { isMounted, styles } = useTransitionStyles(context, {
    duration: 150,
    initial: { opacity: 0, transform: "translateY(0)" },
    open: { opacity: 1, transform: "translateY(-5px)" },
  });

  useEffect(() => {
    if (open) {
      const timeoutId = setTimeout(() => {
        setOpen(false);
      }, 1000);
      return () => clearTimeout(timeoutId);
    }
  }, [open]);

  const handleCopy = useCallback(async () => {
    await navigator.clipboard.writeText(value);
    setOpen(true);
  }, [value, setOpen]);

  return (
    <>
      <CopyIcon
        ref={refs.setReference}
        onClick={handleCopy}
        className="cursor-pointer text-muted-foreground hover:text-foreground h-4 w-4"
      />

      {open && (
        <div ref={refs.setFloating} style={floatingStyles}>
          {isMounted && (
            <div
              style={styles}
              className="font-sans bg-black text-white px-2 py-1 text-xs rounded"
            >
              Copied!
            </div>
          )}
        </div>
      )}
    </>
  );
}

const AuthFormSchema = z.object({
  customAuthDomain: z.string().regex(/[a-z0-9.\-+]/, {
    message:
      "Custom auth.ssoready.com domain must be a DNS hostname, consisting only of a-z, 0-9, dashes, and periods.",
  }),
});

function EditCustomAuthDomainAlertModal({
  customDomainSettings,
}: {
  customDomainSettings: GetEnvironmentCustomDomainSettingsResponse;
}) {
  const { environmentId } = useParams();
  const form = useForm<z.infer<typeof AuthFormSchema>>({
    resolver: zodResolver(AuthFormSchema),
    defaultValues: {
      customAuthDomain: customDomainSettings.customAuthDomain,
    },
  });

  const [open, setOpen] = useState(false);
  const updateCustomDomainSettingsMutation = useMutation(
    updateEnvironmentCustomDomainSettings,
  );
  const queryClient = useQueryClient();
  const handleSubmit = async (
    values: z.infer<typeof AuthFormSchema>,
    e: any,
  ) => {
    e.preventDefault();
    await updateCustomDomainSettingsMutation.mutateAsync({
      environmentId,
      customAuthDomain: values.customAuthDomain,
    });

    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey(getEnvironmentCustomDomainSettings, {
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
                Edit custom auth.ssoready.com domain
              </AlertDialogTitle>
            </AlertDialogHeader>

            <div className="my-4 space-y-4">
              <FormField
                control={form.control}
                name="customAuthDomain"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Custom auth.ssoready.com domain</FormLabel>
                    <FormControl>
                      <Input placeholder="auth.yourcompany.com" {...field} />
                    </FormControl>
                    <FormMessage />
                    <FormDescription>
                      A domain you control that you'd like your customers to see
                      instead of auth.ssoready.com.
                    </FormDescription>
                  </FormItem>
                )}
              />
            </div>

            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <Button
                disabled={updateCustomDomainSettingsMutation.isPending}
                type="submit"
                className="flex items-center gap-x-2"
              >
                {updateCustomDomainSettingsMutation.isPending && (
                  <LoaderCircleIcon className="h-4 w-4 animate-spin" />
                )}
                Save
              </Button>
            </AlertDialogFooter>
          </form>
        </Form>
      </AlertDialogContent>
    </AlertDialog>
  );
}

const AdminFormSchema = z.object({
  customAdminDomain: z.string().regex(/[a-z0-9.\-+]/, {
    message:
      "Custom admin.ssoready.com domain must be a DNS hostname, consisting only of a-z, 0-9, dashes, and periods.",
  }),
});

function EditCustomAdminDomainAlertModal({
  customDomainSettings,
}: {
  customDomainSettings: GetEnvironmentCustomDomainSettingsResponse;
}) {
  const { environmentId } = useParams();
  const form = useForm<z.infer<typeof AdminFormSchema>>({
    resolver: zodResolver(AdminFormSchema),
    defaultValues: {
      customAdminDomain: customDomainSettings.customAdminDomain,
    },
  });

  const [open, setOpen] = useState(false);
  const updateCustomDomainSettingsMutation = useMutation(
    updateEnvironmentCustomDomainSettings,
  );
  const queryClient = useQueryClient();
  const handleSubmit = async (
    values: z.infer<typeof AdminFormSchema>,
    e: any,
  ) => {
    e.preventDefault();
    await updateCustomDomainSettingsMutation.mutateAsync({
      environmentId,
      customAdminDomain: values.customAdminDomain,
    });

    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey(getEnvironmentCustomDomainSettings, {
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
                Edit custom admin.ssoready.com domain
              </AlertDialogTitle>
            </AlertDialogHeader>

            <div className="my-4 space-y-4">
              <FormField
                control={form.control}
                name="customAdminDomain"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Custom admin.ssoready.com domain</FormLabel>
                    <FormControl>
                      <Input placeholder="admin.yourcompany.com" {...field} />
                    </FormControl>
                    <FormMessage />
                    <FormDescription>
                      A domain you control that you'd like your customers to see
                      instead of admin.ssoready.com.
                    </FormDescription>
                  </FormItem>
                )}
              />
            </div>

            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <Button
                disabled={updateCustomDomainSettingsMutation.isPending}
                type="submit"
                className="flex items-center gap-x-2"
              >
                {updateCustomDomainSettingsMutation.isPending && (
                  <LoaderCircleIcon className="h-4 w-4 animate-spin" />
                )}
                Save
              </Button>
            </AlertDialogFooter>
          </form>
        </Form>
      </AlertDialogContent>
    </AlertDialog>
  );
}
