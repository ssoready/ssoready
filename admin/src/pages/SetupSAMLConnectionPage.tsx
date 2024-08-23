import React, { ReactNode, useCallback, useEffect, useState } from "react";
// import { Steps } from "@/components/Steps";
import { Link } from "react-router-dom";
import { CheckIcon, ChevronRightIcon, CopyIcon } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useParams } from "react-router";
import { offset, useFloating, useTransitionStyles } from "@floating-ui/react";
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  adminGetSAMLConnection,
  adminParseSAMLMetadata,
  adminUpdateSAMLConnection,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { useQueryClient } from "@tanstack/react-query";

interface IDP {
  id: string;
  displayName: string;
  subtitle: string;
  firstSubStep: string;
  steps: Step[];
}

const IDPS: IDP[] = [
  {
    id: "okta",
    displayName: "Okta",
    subtitle: "Set up a SAML connection with your corporate Okta.",
    firstSubStep: "okta-create-app",
    steps: [
      {
        displayName: "Create application",
      },
      {
        displayName: "Configure application",
      },
      {
        displayName: "Finish creating application",
      },
      {
        displayName: "Copy Metadata URL",
      },
      {
        displayName: "Assign users",
      },
    ],
  },
  {
    id: "google",
    displayName: "Google",
    subtitle: "Set up a SAML connection with your Google Workspace.",
    firstSubStep: "todo",
    steps: [],
  },
  {
    id: "entra",
    displayName: "Microsoft Entra",
    subtitle:
      "Set up a SAML connection with your Microsoft Entra (formerly known as Azure Active Directory).",
    firstSubStep: "todo",
    steps: [],
  },
  {
    id: "other",
    displayName: "Other Identity Provider",
    subtitle: "Set up a SAML connection with any other identity provider.",
    firstSubStep: "todo",
    steps: [],
  },
];

interface SubStep {
  idpId: string;
  step: number;
}

const SUB_STEPS: Record<string, SubStep> = {
  "okta-create-app": {
    idpId: "okta",
    step: 0,
  },
  "okta-configure-app-name": {
    idpId: "okta",
    step: 1,
  },
  "okta-configure-sso-url": {
    idpId: "okta",
    step: 1,
  },
  "okta-configure-audience-uri": {
    idpId: "okta",
    step: 1,
  },
  "okta-finish-creating-app": {
    idpId: "okta",
    step: 2,
  },
  "okta-copy-metadata-url": {
    idpId: "okta",
    step: 3,
  },
};

export function SetupSAMLConnectionPage() {
  const { samlConnectionId, subStepId } = useParams();
  const subStep = subStepId ? SUB_STEPS[subStepId] : undefined;

  return (
    <>
      <nav className="border-b border-gray-200 bg-white">
        {subStep ? (
          <Steps
            steps={IDPS.find((idp) => idp.id === subStep.idpId)!.steps}
            currentStep={subStep.step}
          />
        ) : (
          <div className="mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex h-16 justify-between">
              <div className="flex text-sm items-center"></div>
              <div className="flex text-sm items-center">
                SAML Connection Setup
              </div>
              <div className="flex text-sm items-center">
                <Link to="/">Cancel</Link>
              </div>
            </div>
          </div>
        )}
      </nav>

      <NarrowContainer>
        {subStepId === undefined && (
          <>
            <h1 className="mt-8 text-xl font-semibold">
              What Identity Provider do you use?
            </h1>

            <div className="mt-4 overflow-hidden bg-white shadow sm:rounded-md">
              <div className="divide-y divide-gray-200">
                {IDPS.map((idp) => (
                  <Link
                    to={`/saml/saml-connections/${samlConnectionId}/setup/${idp.firstSubStep}`}
                    key={idp.id}
                    className="px-4 py-4 sm:px-6 text-sm flex items-center justify-between cursor-pointer"
                  >
                    <div>
                      <div className="font-medium text-gray-900">
                        {idp.displayName}
                      </div>
                      <div className="text-gray-500">{idp.subtitle}</div>
                    </div>

                    <div>
                      <ChevronRightIcon className="text-gray-700" />
                    </div>
                  </Link>
                ))}
              </div>
            </div>
          </>
        )}

        {subStepId === "okta-create-app" && <OktaCreateAppStep />}
        {subStepId === "okta-configure-app-name" && (
          <OktaConfigureAppNameStep />
        )}
        {subStepId === "okta-configure-sso-url" && <OktaConfigureSSOURLStep />}
        {subStepId === "okta-configure-audience-uri" && (
          <OktaConfigureAudienceURIStep />
        )}
        {subStepId === "okta-finish-creating-app" && (
          <OktaFinishCreatingAppStep />
        )}
        {subStepId === "okta-copy-metadata-url" && <OktaCopyMetadataURLStep />}
      </NarrowContainer>
    </>
  );
}

function NarrowContainer({ children }: { children?: ReactNode }) {
  return (
    <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-3xl mt-8 pb-8">{children}</div>
    </div>
  );
}

interface Step {
  displayName: string;
}

function Steps({ steps, currentStep }: { steps: Step[]; currentStep: number }) {
  return (
    <div className="flex w-full justify-center">
      <div className="flex items-center">
        {steps.map((step, index) => (
          <div key={index}>
            {index < currentStep && (
              <div className="group flex w-full items-center">
                <span className="flex items-center px-6 py-4 text-sm font-medium">
                  <span className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-indigo-600 group-hover:bg-indigo-800">
                    <CheckIcon className="h-6 w-6 text-white" />
                  </span>
                  <span className="ml-4 text-sm font-medium text-gray-900">
                    {step.displayName}
                  </span>
                </span>
              </div>
            )}

            {index === currentStep && (
              <div
                aria-current="step"
                className="flex items-center px-6 py-4 text-sm font-medium"
              >
                <span className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full border-2 border-indigo-600">
                  <span className="text-indigo-600">{index + 1}</span>
                </span>
                <span className="ml-4 text-sm font-medium text-indigo-600">
                  {step.displayName}
                </span>
              </div>
            )}

            {index > currentStep && (
              <div className="group flex items-center">
                <span className="flex items-center px-6 py-4 text-sm font-medium">
                  <span className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full border-2 border-gray-300 group-hover:border-gray-400">
                    <span className="text-gray-500 group-hover:text-gray-900">
                      {index + 1}
                    </span>
                  </span>
                  <span className="ml-4 text-sm font-medium text-gray-500 group-hover:text-gray-900">
                    {step.displayName}
                  </span>
                </span>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

function ValueCopier({ value }: { value: string }) {
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
    <div
      ref={refs.setReference}
      onClick={handleCopy}
      className="flex select-none cursor-pointer bg-muted font-mono text-xs border border-input rounded-md px-3 py-2"
    >
      <span>{value}</span>
      <span className="ml-auto flex gap-x-2">
        <CopyIcon className="cursor-pointer text-muted-foreground hover:text-foreground h-4 w-4" />
      </span>

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
    </div>
  );
}

function useSubStepUrl(subStepId: string) {
  const { samlConnectionId } = useParams();
  return `/saml/saml-connections/${samlConnectionId}/setup/${subStepId}`;
}

function OktaCreateAppStep() {
  const next = useSubStepUrl("okta-configure-app-name");
  return (
    <Card>
      <CardHeader>
        <CardTitle>Create an Okta application</CardTitle>
        <CardDescription>
          Create a new Okta application that will let you log into our
          application.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/okta_create_app.gif" />

        <div className="text-sm mt-4">
          <p>Create a new Okta application:</p>

          <ol className="mt-2 list-decimal list-inside space-y-1">
            <li>Go to Applications &gt; Applications in the sidebar.</li>
            <li>Click "Create App Integration"</li>
            <li>Choose "SAML 2.0"</li>
            <li>Click "Next"</li>
          </ol>
        </div>

        <div className="flex justify-end">
          <Button asChild>
            <Link to={next}>Next: Configure application</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function OktaConfigureAppNameStep() {
  const next = useSubStepUrl("okta-configure-sso-url");

  return (
    <Card>
      <CardHeader>
        <CardTitle>Configure app name</CardTitle>
        <CardDescription>
          Configure the name of your Okta app. This is shown to employees.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/okta_configure_app_name.gif" />

        <div className="text-sm mt-4">
          <p>Give the new Okta application a name.</p>
        </div>

        <div className="flex justify-end">
          <Button>
            <Link to={next}>Next: Configure SAML Single Sign-on URL</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function OktaConfigureSSOURLStep() {
  const { samlConnectionId } = useParams();
  const next = useSubStepUrl("okta-configure-audience-uri");
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Configure Single Sign-On URL</CardTitle>
        <CardDescription>
          Configure your app's Single Sign-On URL.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/okta_configure_sso_url.gif" />

        <div className="text-sm mt-4 mb-4">
          <p>Update the "Single sign-on URL" to:</p>
        </div>

        {samlConnection && (
          <ValueCopier value={samlConnection.samlConnection!.spAcsUrl} />
        )}

        <p className="text-sm mt-4">
          Make sure "Use this for Recipient URL and Destination URL" stays
          checked.
        </p>

        <div className="mt-4 flex justify-end">
          <Button>
            <Link to={next}>Next: Configure SAML Audience URI</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function OktaConfigureAudienceURIStep() {
  const { samlConnectionId } = useParams();
  const next = useSubStepUrl("okta-finish-creating-app");
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Configure Audience URI (SP Entity ID)</CardTitle>
        <CardDescription>
          Configure your app's Audience URI (SP Entity ID).
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/okta_configure_audience_uri.gif" />

        <div className="text-sm mt-4 mb-4">
          <p>Update the "Audience URI (SP Entity ID)" to:</p>
        </div>

        {samlConnection && (
          <ValueCopier value={samlConnection.samlConnection!.spEntityId} />
        )}

        <p className="text-sm mt-4">
          No need to touch any other of the SAML settings. Keep them to the
          default values Okta chooses.
        </p>

        <div className="mt-4 flex justify-end">
          <Button>
            <Link to={next}>Next: Finish Creating App</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function OktaFinishCreatingAppStep() {
  const next = useSubStepUrl("okta-copy-metadata-url");

  return (
    <Card>
      <CardHeader>
        <CardTitle>Finish creating app</CardTitle>
        <CardDescription>Finish creating your Okta app.</CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/okta_finish_creating_app.gif" />

        <div className="text-sm mt-4">
          <p>Finish creating your Okta app.</p>

          <ol className="mt-2 list-decimal list-inside space-y-1">
            <li>Click "Next".</li>
            <li>Click "This is an internal app that we have created"</li>
            <li>Click "Finish"</li>
          </ol>
        </div>

        <div className="flex justify-end">
          <Button>
            <Link to={next}>Next: Copy Metadata URL</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

const oktaMetadataURLFormSchema = z.object({
  metadataUrl: z
    .string()
    .min(1, { message: "SAML Metadata URL is required." })
    .url({ message: "SAML Metadata URL must be a valid URL." }),
});

function OktaCopyMetadataURLStep() {
  const { samlConnectionId } = useParams();
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });
  const next = useSubStepUrl("okta-assign-users");

  const [success, setSuccess] = useState(false);

  const form = useForm<z.infer<typeof oktaMetadataURLFormSchema>>({
    resolver: zodResolver(oktaMetadataURLFormSchema),
    defaultValues: {
      metadataUrl: "",
    },
  });

  const parseSAMLMetadataMutation = useMutation(adminParseSAMLMetadata);
  const updateSAMLConnectionMutation = useMutation(adminUpdateSAMLConnection);
  const queryClient = useQueryClient();
  const handleSubmit = async (
    data: z.infer<typeof oktaMetadataURLFormSchema>,
    e: any,
  ) => {
    e.preventDefault();

    const { idpRedirectUrl, idpCertificate, idpEntityId } =
      await parseSAMLMetadataMutation.mutateAsync({ url: data.metadataUrl });

    await updateSAMLConnectionMutation.mutateAsync({
      samlConnection: {
        id: samlConnection!.samlConnection!.id,
        primary: samlConnection!.samlConnection!.primary,
        idpRedirectUrl,
        idpCertificate,
        idpEntityId,
      },
    });

    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey(adminGetSAMLConnection, {
        id: samlConnection!.samlConnection!.id,
      }),
    });

    setSuccess(true);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Copy SAML metadata URL</CardTitle>
        <CardDescription>Copy your app's SAML metadata URL.</CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/okta_copy_metadata_url.gif" />

        {!success && (
          <>
            <div className="text-sm mt-4">
              <p>Copy your app's SAML metadata URL.</p>

              <ol className="mt-2 list-decimal list-inside space-y-1">
                <li>Click on the "Sign On" tab</li>
                <li>Scroll down to where you see "Metadata URL"</li>
                <li>Click "Copy"</li>
              </ol>

              <p className="mt-2">Paste the SAML metadata URL here:</p>
            </div>

            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(handleSubmit)}
                className="mt-4 w-full space-y-6"
              >
                <FormField
                  control={form.control}
                  name="metadataUrl"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>SAML Metadata URL</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormDescription>
                        You should copy-paste this value from Okta.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="flex justify-end">
                  <Button disabled={success}>Submit</Button>
                </div>
              </form>
            </Form>
          </>
        )}

        {success && (
          <div>
            <p className="text-sm mt-4">
              Successfully imported your app's SAML settings from Okta. The last
              remaining step is to assign users to your Okta app.
            </p>

            <div className="flex justify-end">
              <Button asChild>
                <Link to={next}>Next: Assign users</Link>
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
