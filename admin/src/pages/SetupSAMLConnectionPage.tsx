import React, { ReactNode, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import {
  CheckIcon,
  ChevronDownIcon,
  ChevronRightIcon,
} from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useParams } from "react-router";
import { ValueCopier } from "@/components/ValueCopier";
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  adminCreateTestModeSAMLFlow,
  adminGetSAMLConnection,
  adminParseSAMLMetadata,
  adminUpdateSAMLConnection,
  adminWhoami,
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
import { useTitle } from "@/useTitle";
import { Helmet } from "react-helmet";
import { Fireworks } from "fireworks-js/dist/react";

interface IDP {
  id: string;
  displayName: string;
  logoUrl: string;
  subtitle: string;
  firstSubStep: string;
  steps: Step[];
}

const IDPS: IDP[] = [
  {
    id: "okta",
    displayName: "Okta",
    logoUrl: "/logo-okta.svg",
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
    logoUrl: "/logo-google.svg",
    subtitle: "Set up a SAML connection with your Google Workspace.",
    firstSubStep: "google-create-app",
    steps: [
      {
        displayName: "Create application",
      },
      {
        displayName: "Download metadata",
      },
      {
        displayName: "Configure application",
      },
      {
        displayName: "Assign users",
      },
    ],
  },
  {
    id: "entra",
    displayName: "Microsoft Entra",
    logoUrl: "/logo-entra.svg",
    subtitle:
      "Set up a SAML connection with your Microsoft Entra (formerly known as Azure Active Directory).",
    firstSubStep: "entra-create-app",
    steps: [
      {
        displayName: "Create application",
      },
      {
        displayName: "Configure application",
      },
      {
        displayName: "Download metadata",
      },
      {
        displayName: "Assign users",
      },
    ],
  },
  {
    id: "other",
    displayName: "Other Identity Provider",
    logoUrl: "/logo-other.svg",
    subtitle: "Set up a SAML connection with any other identity provider.",
    firstSubStep: "other-create-app",
    steps: [
      {
        displayName: "Create application",
      },
      {
        displayName: "Configure application",
      },
      {
        displayName: "Download metadata",
      },
      {
        displayName: "Assign users",
      },
    ],
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
  "okta-assign-users": {
    idpId: "okta",
    step: 4,
  },
  "okta-test-success": {
    idpId: "okta",
    step: 5,
  },
  "okta-complete": {
    idpId: "okta",
    step: 5,
  },
  "google-create-app": {
    idpId: "google",
    step: 0,
  },
  "google-configure-app-name": {
    idpId: "google",
    step: 0,
  },
  "google-download-metadata": {
    idpId: "google",
    step: 1,
  },
  "google-configure-acs-url": {
    idpId: "google",
    step: 2,
  },
  "google-configure-entity-id": {
    idpId: "google",
    step: 2,
  },
  "google-assign-users": {
    idpId: "google",
    step: 3,
  },
  "google-test-success": {
    idpId: "google",
    step: 4,
  },
  "google-complete": {
    idpId: "google",
    step: 4,
  },
  "entra-create-app": {
    idpId: "entra",
    step: 0,
  },
  "entra-configure-entity-id": {
    idpId: "entra",
    step: 1,
  },
  "entra-configure-reply-url": {
    idpId: "entra",
    step: 1,
  },
  "entra-download-metadata": {
    idpId: "entra",
    step: 2,
  },
  "entra-assign-users": {
    idpId: "entra",
    step: 3,
  },
  "entra-test-success": {
    idpId: "entra",
    step: 4,
  },
  "entra-complete": {
    idpId: "entra",
    step: 4,
  },
  "other-create-app": {
    idpId: "other",
    step: 0,
  },
  "other-configure-app": {
    idpId: "other",
    step: 1,
  },
  "other-download-metadata": {
    idpId: "other",
    step: 2,
  },
  "other-assign-users": {
    idpId: "other",
    step: 3,
  },
  "other-test-success": {
    idpId: "other",
    step: 4,
  },
  "other-complete": {
    idpId: "other",
    step: 4,
  },
};

export function SetupSAMLConnectionPage() {
  const title = useTitle("SAML Connection Setup");
  const { samlConnectionId, subStepId } = useParams();
  const subStep = subStepId ? SUB_STEPS[subStepId] : undefined;

  return (
    <>
      <Helmet>
        <title>{title}</title>
      </Helmet>
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
                    className="py-4 pl-4 pr-6 text-sm flex items-center justify-between cursor-pointer"
                  >
                    <div className="flex items-center gap-x-4">
                      <img className="h-8 w-8" src={idp.logoUrl} alt="" />
                      <div>
                        <div className="font-medium text-gray-900">
                          {idp.displayName}
                        </div>
                        <div className="text-gray-500">{idp.subtitle}</div>
                      </div>
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
        {subStepId === "okta-assign-users" && <OktaAssignUsersStep />}
        {subStepId === "okta-test-success" && <OktaTestSuccessPage />}
        {subStepId === "okta-complete" && <CompleteStep />}

        {subStepId === "google-create-app" && <GoogleCreateAppStep />}
        {subStepId === "google-configure-app-name" && (
          <GoogleConfigureAppNameStep />
        )}
        {subStepId === "google-download-metadata" && (
          <GoogleDownloadMetadataStep />
        )}
        {subStepId === "google-configure-acs-url" && (
          <GoogleConfigureACSURLStep />
        )}
        {subStepId === "google-configure-entity-id" && (
          <GoogleConfigureEntityIDStep />
        )}
        {subStepId === "google-assign-users" && <GoogleAssignUsersStep />}
        {subStepId === "google-test-success" && <GoogleTestSuccessPage />}
        {subStepId === "google-complete" && <CompleteStep />}

        {subStepId === "entra-create-app" && <EntraCreateAppStep />}
        {subStepId === "entra-configure-entity-id" && (
          <EntraConfigureEntityIDStep />
        )}
        {subStepId === "entra-configure-reply-url" && (
          <EntraConfigureReplyURLStep />
        )}
        {subStepId === "entra-download-metadata" && (
          <EntraDownloadMetadataStep />
        )}
        {subStepId === "entra-assign-users" && <EntraAssignUsersStep />}
        {subStepId === "entra-test-success" && <EntraTestSuccessPage />}
        {subStepId === "entra-complete" && <CompleteStep />}

        {subStepId === "other-create-app" && <OtherCreateAppStep />}
        {subStepId === "other-configure-app" && <OtherConfigureAppStep />}
        {subStepId === "other-download-metadata" && (
          <OtherDownloadMetadataStep />
        )}
        {subStepId === "other-assign-users" && <OtherAssignUsersStep />}
        {subStepId === "other-test-success" && <OtherTestSuccessPage />}
        {subStepId === "other-complete" && <CompleteStep />}
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
                  <span className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full border-2 border-indigo-600">
                    <CheckIcon className="h-6 w-6 text-indigo-600" />
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
                  <span className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full border-2 border-gray-300">
                    <span className="text-gray-500">{index + 1}</span>
                  </span>
                  <span className="ml-4 text-sm font-medium text-gray-500">
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

function OktaAssignUsersStep() {
  const { samlConnectionId } = useParams();
  const next = useSubStepUrl("okta-complete");

  const createTestModeSAMLFlowMutation = useMutation(
    adminCreateTestModeSAMLFlow,
  );

  const handleTest = async () => {
    const { redirectUrl } = await createTestModeSAMLFlowMutation.mutateAsync({
      samlConnectionId: samlConnectionId,
      testModeIdp: "okta",
    });

    location.href = redirectUrl;
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Assign users to app</CardTitle>
        <CardDescription>Assign users to your new app.</CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/okta_assign_users.gif" />

        <div className="text-sm mt-4">
          <p>
            Assign users to the new app. If you intend to test the connection
            yourself, remember to assign yourself too.
          </p>

          <ol className="mt-2 list-decimal list-inside space-y-1">
            <li>Click on the "Assignments" tab</li>
            <li>Click "Assign"</li>
            <li>
              Click "Assign to People" or "Assign to Groups", whichever you
              usually use
            </li>
          </ol>

          <p className="mt-2">
            Once you've assigned the appropriate users to the app, you're done
            setting up SAML.
          </p>
        </div>

        <div className="mt-4 flex gap-x-2 justify-end">
          <Button variant="secondary">
            <Link to={next}>Skip testing</Link>
          </Button>
          <Button onClick={handleTest}>Test SAML connection</Button>
        </div>
      </CardContent>
    </Card>
  );
}

function OktaTestSuccessPage() {
  const next = useSubStepUrl("okta-complete");
  const edit = useSubStepUrl("okta-configure-sso-url");
  const [searchParams] = useSearchParams();
  const email = searchParams.get("email")!;
  const attributes = JSON.parse(searchParams.get("attributes")!) as Record<
    string,
    string
  >;

  return (
    <Card>
      <CardHeader>
        <CardTitle>SAML login successful</CardTitle>
        <CardDescription>Here's what we received from Okta</CardDescription>
      </CardHeader>

      <CardContent>
        <div className="text-sm">
          <p className="font-semibold text-muted-foreground">
            Email (Subject ID)
          </p>
          <p className="mt-1 font-medium">{email}</p>

          <p className="mt-4 font-semibold text-muted-foreground">Attributes</p>
          <div className="mt-2 flex flex-col gap-y-2">
            {Object.entries(attributes).map(([key, value]) => (
              <div key={key} className="bg-muted p-2 rounded-md">
                <p className="font-medium text-muted-foreground">{key}</p>
                <p className="font-semibold">{value}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-4 flex gap-x-2 justify-end">
          <Button variant="secondary" asChild>
            <Link to={edit}>Edit connection</Link>
          </Button>
          <Button asChild>
            <Link to={next}>Setup compete!</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function GoogleCreateAppStep() {
  const next = useSubStepUrl("google-configure-app-name");
  return (
    <Card>
      <CardHeader>
        <CardTitle>Create a Google SAML application</CardTitle>
        <CardDescription>
          Create a new Google SAML application that will let you log into our
          application.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/google_create_app.gif" />

        <div className="text-sm mt-4">
          <p>Create a new Google SAML application:</p>

          <ol className="mt-2 list-decimal list-inside space-y-1">
            <li>Go to admin.google.com.</li>
            <li>In the sidebar, click "Apps &gt; Web and mobile apps"</li>
            <li>Click on the "Add app" dropdown</li>
            <li>Click "Add custom SAML app"</li>
          </ol>
        </div>

        <div className="flex justify-end">
          <Button asChild>
            <Link to={next}>Next: Configure application name</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function GoogleConfigureAppNameStep() {
  const next = useSubStepUrl("google-download-metadata");
  return (
    <Card>
      <CardHeader>
        <CardTitle>Configure app name</CardTitle>
        <CardDescription>
          Configure the name of your Google app. This is shown to employees.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/google_configure_app_name.gif" />

        <div className="text-sm mt-4 mb-4">
          <p>
            Give the new Google application a name. Optionally, provide a
            description and upload a logo. Then click "Continue".
          </p>
        </div>

        <div className="flex justify-end">
          <Button>
            <Link to={next}>Next: Download Metadata</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

const googleMetadataFormSchema = z.object({
  metadata: z
    .string()
    .min(1, { message: "GoogleIDPMetadata.xml is required." }),
});

function GoogleDownloadMetadataStep() {
  const { samlConnectionId } = useParams();
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });
  const next = useSubStepUrl("google-configure-acs-url");

  const [success, setSuccess] = useState(false);

  const form = useForm<z.infer<typeof googleMetadataFormSchema>>({
    resolver: zodResolver(googleMetadataFormSchema),
    defaultValues: {
      metadata: "",
    },
  });

  const parseSAMLMetadataMutation = useMutation(adminParseSAMLMetadata);
  const updateSAMLConnectionMutation = useMutation(adminUpdateSAMLConnection);
  const queryClient = useQueryClient();
  const handleSubmit = async (
    data: z.infer<typeof googleMetadataFormSchema>,
    e: any,
  ) => {
    e.preventDefault();

    const { idpRedirectUrl, idpCertificate, idpEntityId } =
      await parseSAMLMetadataMutation.mutateAsync({ xml: data.metadata });

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
        <CardTitle>Download IdP metadata</CardTitle>
        <CardDescription>
          Download your new application's IdP metadata, and upload it here.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/google_download_metadata.gif" />

        {!success && (
          <>
            <div className="text-sm mt-4 mb-4">
              <p>
                Click on "Download Metadata". You'll download a
                GoogleIDPMetadata.xml file. Upload that file here.
              </p>
            </div>

            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(handleSubmit)}
                className="mt-4 w-full space-y-6"
              >
                <FormField
                  control={form.control}
                  name="metadata"
                  render={({ field: { onChange } }) => (
                    <FormItem>
                      <FormLabel>GoogleIDPMetadata.xml</FormLabel>
                      <FormControl>
                        <Input
                          type="file"
                          onChange={async (e) => {
                            // File inputs are special; they are necessarily "uncontrolled", and their value is a FileList.
                            // We just copy over the file's contents to the react-form-hook state manually on input change.
                            if (e.target.files) {
                              onChange(await e.target.files[0].text());
                            }
                          }}
                        />
                      </FormControl>
                      <FormDescription>
                        You should upload the GoogleIDPMetadata.xml file you
                        downloaded from Google.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="flex justify-end">
                  <Button>Submit</Button>
                </div>
              </form>
            </Form>
          </>
        )}

        {success && (
          <div>
            <p className="text-sm mt-4">
              Successfully imported your app's SAML settings from Google. Now
              click "Continue" in Google, if you haven't already. You should see
              a screen titled "Service provider details".
            </p>

            <div className="flex justify-end">
              <Button asChild>
                <Link to={next}>Next: Configure app</Link>
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function GoogleConfigureACSURLStep() {
  const { samlConnectionId } = useParams();
  const next = useSubStepUrl("google-configure-entity-id");
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Configure ACS URL</CardTitle>
        <CardDescription>Configure your app's ACS URL.</CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/google_configure_acs_url.gif" />

        <div className="text-sm mt-4 mb-4">
          <p>Update the "ACS URL" to:</p>
        </div>

        {samlConnection && (
          <ValueCopier value={samlConnection.samlConnection!.spAcsUrl} />
        )}

        <div className="mt-4 flex justify-end">
          <Button>
            <Link to={next}>Next: Assign Users</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function GoogleConfigureEntityIDStep() {
  const { samlConnectionId } = useParams();
  const next = useSubStepUrl("google-assign-users");
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Configure Entity ID</CardTitle>
        <CardDescription>Configure your app's Entity ID.</CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/google_configure_entity_id.gif" />

        <div className="text-sm mt-4 mb-4">
          <p>Update the "Entity ID" to:</p>
        </div>

        {samlConnection && (
          <ValueCopier value={samlConnection.samlConnection!.spEntityId} />
        )}

        <p className="text-sm mt-4">
          Then click "Continue", and then "Finish". You have now configured the
          SAML application. The last step is to assign users to the application.
        </p>

        <div className="mt-4 flex justify-end">
          <Button>
            <Link to={next}>Next: Assign Users</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function GoogleAssignUsersStep() {
  const next = useSubStepUrl("google-complete");
  const { samlConnectionId } = useParams();
  const createTestModeSAMLFlowMutation = useMutation(
    adminCreateTestModeSAMLFlow,
  );

  const handleTest = async () => {
    const { redirectUrl } = await createTestModeSAMLFlowMutation.mutateAsync({
      samlConnectionId: samlConnectionId,
      testModeIdp: "google",
    });

    location.href = redirectUrl;
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Assign users to app</CardTitle>
        <CardDescription>Assign users to your new app.</CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/google_assign_users.gif" />

        <div className="text-sm mt-4">
          <p>
            Assign users to the new app. If you're familiar with Google
            Workspace organizational units, use whatever process you normally
            use.
          </p>

          <p className="mt-4">
            If you're not familiar with Google Workspace organizational units,
            or you don't normally use them, here's the simplest way to assign
            users to your new Google Workspace SAML application:
          </p>

          <ol className="mt-2 list-decimal list-inside space-y-1">
            <li>
              To the right of "User access" is a chevron pointing down (
              <ChevronDownIcon className="inline h-4 w-4" />
              ). Click on it.
            </li>
            <li>Click on "ON for everyone"</li>
            <li>Click "Save".</li>
          </ol>

          <p className="mt-2">
            Allow a minute before users, including yourself, can log in. Google
            Workspace doesn't immediately reflect permissions updates.
          </p>
        </div>

        <div className="mt-4 flex gap-x-2 justify-end">
          <Button variant="secondary">
            <Link to={next}>Skip testing</Link>
          </Button>
          <Button onClick={handleTest}>Test SAML connection</Button>
        </div>
      </CardContent>
    </Card>
  );
}

function GoogleTestSuccessPage() {
  const next = useSubStepUrl("google-complete");
  const edit = useSubStepUrl("google-create-app");
  const [searchParams] = useSearchParams();
  const email = searchParams.get("email")!;
  const attributes = JSON.parse(searchParams.get("attributes")!) as Record<
    string,
    string
  >;

  return (
    <Card>
      <CardHeader>
        <CardTitle>SAML login successful</CardTitle>
        <CardDescription>Here's what we received from Google</CardDescription>
      </CardHeader>

      <CardContent>
        <div className="text-sm">
          <p className="font-semibold text-muted-foreground">
            Email (Subject ID)
          </p>
          <p className="mt-1 font-medium">{email}</p>

          <p className="mt-4 font-semibold text-muted-foreground">Attributes</p>
          <div className="mt-2 flex flex-col gap-y-2">
            {Object.entries(attributes).map(([key, value]) => (
              <div key={key} className="bg-muted p-2 rounded-md">
                <p className="font-medium text-muted-foreground">{key}</p>
                <p className="font-semibold">{value}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-4 flex gap-x-2 justify-end">
          <Button variant="secondary" asChild>
            <Link to={edit}>Edit connection</Link>
          </Button>
          <Button asChild>
            <Link to={next}>Setup compete!</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function EntraCreateAppStep() {
  const next = useSubStepUrl("entra-configure-entity-id");
  return (
    <Card>
      <CardHeader>
        <CardTitle>Create an Entra application</CardTitle>
        <CardDescription>
          Create a new Entra application that will let you log into our
          application.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/entra_create_app.gif" />

        <div className="text-sm mt-4">
          <p>Create a new Entra application:</p>

          <ol className="mt-2 list-decimal list-inside space-y-1">
            <li>Go to entra.microsoft.com.</li>
            <li>
              In the sidebar, click "Applications &gt; Enterprise Applications"
            </li>
            <li>Click on the "New application"</li>
            <li>Click "Create your own application"</li>
            <li>Enter a name into "What's the name of your app?"</li>
            <li>
              Keep "Integrate any other application you don't find in the
              gallery (Non-gallery)" checked.
            </li>
            <li>Click "Create"</li>
          </ol>
        </div>

        <div className="mt-4 flex justify-end">
          <Button asChild>
            <Link to={next}>Next: Configure application</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function EntraConfigureEntityIDStep() {
  const { samlConnectionId } = useParams();
  const next = useSubStepUrl("entra-configure-reply-url");
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Configure SAML Identifier (Entity ID)</CardTitle>
        <CardDescription>
          Configure your app's SAML Identifier (Entity ID).
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/entra_configure_entity_id.gif" />

        <div className="text-sm my-4">
          <p>Update your Entra application's SAML Identifier (Entity ID):</p>

          <ol className="mt-2 list-decimal list-inside space-y-1">
            <li>Navigate to your application if you haven't already.</li>
            <li>In the sidebar for the application, click "Single sign-on"</li>
            <li>Click on "SAML"</li>
            <li>
              Click on "Edit" icon to the right of "Basic SAML Configuration"
            </li>
            <li>
              Click "Add identifier". An input now appears under the "Identifier
              (Entity ID)" section.
            </li>
            <li>Enter the following value:</li>
          </ol>
        </div>

        {samlConnection && (
          <ValueCopier value={samlConnection.samlConnection!.spEntityId} />
        )}

        <p className="text-sm mt-4">
          No need to click "Add identifier" again. Once you've pasted the value
          into the input, move to the next step.
        </p>

        <div className="mt-4 flex justify-end">
          <Button>
            <Link to={next}>Next: Configure Reply URL</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function EntraConfigureReplyURLStep() {
  const { samlConnectionId } = useParams();
  const next = useSubStepUrl("entra-download-metadata");
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          Configure Reply URL (Assertion Consumer Service URL)
        </CardTitle>
        <CardDescription>
          Configure your app's Reply URL (Assertion Consumer Service URL).
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/entra_configure_reply_url.gif" />

        <div className="text-sm mt-4 mb-4">
          <p>
            Under the "Reply URL (Assertion Consumer Service URL)" section,
            click "Add reply URL". An input now appears. Paste in this value:
          </p>
        </div>

        {samlConnection && (
          <ValueCopier value={samlConnection.samlConnection!.spAcsUrl} />
        )}

        <p className="my-4 text-sm">
          Keep all other settings to their default values. Click "Save" above.
        </p>

        <div className="mt-4 flex justify-end">
          <Button>
            <Link to={next}>Next: Download metadata</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

const entraMetadataFormSchema = z.object({
  metadata: z
    .string()
    .min(1, { message: "Federation Metadata XML file is required." }),
});

function EntraDownloadMetadataStep() {
  const { samlConnectionId } = useParams();
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });
  const next = useSubStepUrl("entra-assign-users");

  const [success, setSuccess] = useState(false);

  const form = useForm<z.infer<typeof entraMetadataFormSchema>>({
    resolver: zodResolver(entraMetadataFormSchema),
    defaultValues: {
      metadata: "",
    },
  });

  const parseSAMLMetadataMutation = useMutation(adminParseSAMLMetadata);
  const updateSAMLConnectionMutation = useMutation(adminUpdateSAMLConnection);
  const queryClient = useQueryClient();
  const handleSubmit = async (
    data: z.infer<typeof entraMetadataFormSchema>,
    e: any,
  ) => {
    e.preventDefault();

    const { idpRedirectUrl, idpCertificate, idpEntityId } =
      await parseSAMLMetadataMutation.mutateAsync({ xml: data.metadata });

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
        <CardTitle>Download Federation Metadata XML</CardTitle>
        <CardDescription>
          Download your new application's Federation Metadata XML, and upload it
          here.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/entra_download_metadata.gif" />

        {!success && (
          <>
            <div className="text-sm my-4">
              <p>Download your application's Federation Metadata XML:</p>

              <ol className="mt-2 list-decimal list-inside space-y-1">
                <li>Navigate to your application if you haven't already.</li>
                <li>
                  In the sidebar for your application, click "Single sign-On"
                </li>
                <li>Scroll down to the "SAML Certificates" section</li>
                <li>
                  Click on the "Download" link next to "Federation Metadata XML"
                </li>
                <li>
                  Your browser now downloads a file. Upload that file here:
                </li>
              </ol>
            </div>

            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(handleSubmit)}
                className="mt-4 w-full space-y-6"
              >
                <FormField
                  control={form.control}
                  name="metadata"
                  render={({ field: { onChange } }) => (
                    <FormItem>
                      <FormLabel>Federation Metadata XML</FormLabel>
                      <FormControl>
                        <Input
                          type="file"
                          onChange={async (e) => {
                            // File inputs are special; they are necessarily "uncontrolled", and their value is a FileList.
                            // We just copy over the file's contents to the react-form-hook state manually on input change.
                            if (e.target.files) {
                              onChange(await e.target.files[0].text());
                            }
                          }}
                        />
                      </FormControl>
                      <FormDescription>
                        You should upload the file you downloaded from Entra. It
                        should be a .xml file named after your Entra
                        application.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="flex justify-end">
                  <Button>Submit</Button>
                </div>
              </form>
            </Form>
          </>
        )}

        {success && (
          <div>
            <p className="text-sm mt-4">
              Successfully imported your app's Federation Metadata XML settings
              from Entra. The last remaining step is to assign users to your new
              application.
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

function EntraAssignUsersStep() {
  const next = useSubStepUrl("entra-complete");
  const { samlConnectionId } = useParams();
  const createTestModeSAMLFlowMutation = useMutation(
    adminCreateTestModeSAMLFlow,
  );

  const handleTest = async () => {
    const { redirectUrl } = await createTestModeSAMLFlowMutation.mutateAsync({
      samlConnectionId: samlConnectionId,
      testModeIdp: "entra",
    });

    location.href = redirectUrl;
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Assign users to app</CardTitle>
        <CardDescription>Assign users to your new app.</CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/entra_assign_users.gif" />

        <div className="text-sm mt-4">
          <p>
            Assign users to the new app. If you're familiar with Entra
            application user assignments, use whatever process you normally use.
          </p>

          <p className="mt-4">
            Otherwise, the most straightforward process is to:
          </p>

          <ol className="mt-2 list-decimal list-inside space-y-1">
            <li>Click on "Users and groups" in the application sidebar</li>
            <li>Click on "Add user/group"</li>
            <li>Under "Users", click on "None Selected"</li>
            <li>
              Check the checkbox next to each of the users you want to assign to
              the application. If you intend to test the application yourself,
              remember to include yourself
            </li>
            <li>Click "Select" at the bottom</li>
            <li>Click "Assign"</li>
          </ol>

          <p className="mt-2">Your application is now configured.</p>
        </div>

        <div className="mt-4 flex gap-x-2 justify-end">
          <Button variant="secondary">
            <Link to={next}>Skip testing</Link>
          </Button>
          <Button onClick={handleTest}>Test SAML connection</Button>
        </div>
      </CardContent>
    </Card>
  );
}

function EntraTestSuccessPage() {
  const next = useSubStepUrl("entra-complete");
  const edit = useSubStepUrl("entra-create-app");
  const [searchParams] = useSearchParams();
  const email = searchParams.get("email")!;
  const attributes = JSON.parse(searchParams.get("attributes")!) as Record<
    string,
    string
  >;

  return (
    <Card>
      <CardHeader>
        <CardTitle>SAML login successful</CardTitle>
        <CardDescription>Here's what we received from Entra</CardDescription>
      </CardHeader>

      <CardContent>
        <div className="text-sm">
          <p className="font-semibold text-muted-foreground">
            Email (Subject ID)
          </p>
          <p className="mt-1 font-medium">{email}</p>

          <p className="mt-4 font-semibold text-muted-foreground">Attributes</p>
          <div className="mt-2 flex flex-col gap-y-2">
            {Object.entries(attributes).map(([key, value]) => (
              <div key={key} className="bg-muted p-2 rounded-md">
                <p className="font-medium text-muted-foreground">{key}</p>
                <p className="font-semibold">{value}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-4 flex gap-x-2 justify-end">
          <Button variant="secondary" asChild>
            <Link to={edit}>Edit connection</Link>
          </Button>
          <Button asChild>
            <Link to={next}>Setup compete!</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function OtherCreateAppStep() {
  const next = useSubStepUrl("other-configure-app");

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create a SAML application</CardTitle>
        <CardDescription>
          Create a new SAML application inside your Identity Provider.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <p className="text-sm">
          Inside your Identity Provider, create a new SAML application.
        </p>

        <div className="mt-4 flex justify-end">
          <Button asChild>
            <Link to={next}>Next: Configure application</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function OtherConfigureAppStep() {
  const { samlConnectionId } = useParams();
  const next = useSubStepUrl("other-download-metadata");
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Configure SAML application</CardTitle>
        <CardDescription>
          Configure your SAML application's Service Provider settings
        </CardDescription>
      </CardHeader>

      <CardContent>
        <p className="text-sm">
          Inside your Identity Provider, configure your application's service
          provider settings.
        </p>

        <p className="text-sm my-4">
          Your service provider will ask for a Service Provider ACS URL, or some
          variation of one of these names:
        </p>

        <ul className="text-sm mt-2 list-disc list-inside space-y-1">
          <li>Assertion Consumer Service URL</li>
          <li>SAML Start URL</li>
          <li>SAML Sign-On URL</li>
        </ul>

        <p className="text-sm my-4">
          These terms all refer to the same concept. Input the following:
        </p>

        {samlConnection && (
          <ValueCopier value={samlConnection.samlConnection!.spAcsUrl} />
        )}

        <p className="text-sm my-4">
          Your service provider will also ask for a Service Provider Entity ID,
          or some variation of:
        </p>

        <ul className="text-sm mt-2 list-disc list-inside space-y-1">
          <li>SP Entity Identifier</li>
          <li>Audience URI</li>
          <li>Relying Party Identifier / ID</li>
        </ul>

        <p className="text-sm my-4">
          These terms all refer to the same concept. Input the following:
        </p>

        {samlConnection && (
          <ValueCopier value={samlConnection.samlConnection!.spEntityId} />
        )}

        <p className="text-sm my-4">
          After you've configured these values in your Identity Provider, the
          next step is to download your application's SAML metadata.
        </p>

        <div className="mt-4 flex justify-end">
          <Button asChild>
            <Link to={next}>Next: Download metadata</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

const otherMetadataFormSchema = z.object({
  metadata: z
    .string()
    .min(1, { message: "SAML Metadata XML file is required." }),
});

function OtherDownloadMetadataStep() {
  const { samlConnectionId } = useParams();
  const { data: samlConnection } = useQuery(adminGetSAMLConnection, {
    id: samlConnectionId,
  });
  const next = useSubStepUrl("other-assign-users");

  const [success, setSuccess] = useState(false);

  const form = useForm<z.infer<typeof otherMetadataFormSchema>>({
    resolver: zodResolver(otherMetadataFormSchema),
    defaultValues: {
      metadata: "",
    },
  });

  const parseSAMLMetadataMutation = useMutation(adminParseSAMLMetadata);
  const updateSAMLConnectionMutation = useMutation(adminUpdateSAMLConnection);
  const queryClient = useQueryClient();
  const handleSubmit = async (
    data: z.infer<typeof otherMetadataFormSchema>,
    e: any,
  ) => {
    e.preventDefault();

    const { idpRedirectUrl, idpCertificate, idpEntityId } =
      await parseSAMLMetadataMutation.mutateAsync({ xml: data.metadata });

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
        <CardTitle>Download SAML metadata XML</CardTitle>
        <CardDescription>
          Download your new application's SAML metadata XML file.
        </CardDescription>
      </CardHeader>

      <CardContent>
        {!success && (
          <>
            <p className="text-sm">
              Find your new SAML application's metadata XML file. This is a file
              your Identity Provider allows you to download on each of your SAML
              applications. Upload that file here.
            </p>

            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(handleSubmit)}
                className="mt-4 w-full space-y-6"
              >
                <FormField
                  control={form.control}
                  name="metadata"
                  render={({ field: { onChange } }) => (
                    <FormItem>
                      <FormLabel>Identity Provider SAML Metadata XML</FormLabel>
                      <FormControl>
                        <Input
                          type="file"
                          onChange={async (e) => {
                            // File inputs are special; they are necessarily "uncontrolled", and their value is a FileList.
                            // We just copy over the file's contents to the react-form-hook state manually on input change.
                            if (e.target.files) {
                              onChange(await e.target.files[0].text());
                            }
                          }}
                        />
                      </FormControl>
                      <FormDescription>
                        You should upload the file you downloaded from your
                        identity provider. It should be a .xml file.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="flex justify-end">
                  <Button>Submit</Button>
                </div>
              </form>
            </Form>
          </>
        )}

        {success && (
          <div>
            <p className="text-sm">
              Successfully imported your app's SAML metadata. The last remaining
              step is to assign users to your SAML application.
            </p>

            <div className="mt-4 flex justify-end">
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

function OtherAssignUsersStep() {
  const next = useSubStepUrl("other-complete");
  const { samlConnectionId } = useParams();
  const createTestModeSAMLFlowMutation = useMutation(
    adminCreateTestModeSAMLFlow,
  );

  const handleTest = async () => {
    const { redirectUrl } = await createTestModeSAMLFlowMutation.mutateAsync({
      samlConnectionId: samlConnectionId,
      testModeIdp: "other",
    });

    location.href = redirectUrl;
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Assign users to app</CardTitle>
        <CardDescription>Assign users to your new app.</CardDescription>
      </CardHeader>

      <CardContent>
        <div className="text-sm">
          <p>
            Assign users to the new app. In your Identity Provider, go to your
            new SAML application's settings related to user or group
            assignments. Assign the appropriate users to the application. If you
            intend to test the application yourself, remember to include
            yourself.
          </p>

          <p className="mt-2">
            Once you've completed this step, your application is now configured.
          </p>
        </div>

        <div className="mt-4 flex gap-x-2 justify-end">
          <Button variant="secondary">
            <Link to={next}>Skip testing</Link>
          </Button>
          <Button onClick={handleTest}>Test SAML connection</Button>
        </div>
      </CardContent>
    </Card>
  );
}

function OtherTestSuccessPage() {
  const next = useSubStepUrl("other-complete");
  const edit = useSubStepUrl("other-create-app");
  const [searchParams] = useSearchParams();
  const email = searchParams.get("email")!;
  const attributes = JSON.parse(searchParams.get("attributes")!) as Record<
    string,
    string
  >;

  return (
    <Card>
      <CardHeader>
        <CardTitle>SAML login successful</CardTitle>
        <CardDescription>
          Here's what we received from your identity provider
        </CardDescription>
      </CardHeader>

      <CardContent>
        <div className="text-sm">
          <p className="font-semibold text-muted-foreground">
            Email (Subject ID)
          </p>
          <p className="mt-1 font-medium">{email}</p>

          <p className="mt-4 font-semibold text-muted-foreground">Attributes</p>
          <div className="mt-2 flex flex-col gap-y-2">
            {Object.entries(attributes).map(([key, value]) => (
              <div key={key} className="bg-muted p-2 rounded-md">
                <p className="font-medium text-muted-foreground">{key}</p>
                <p className="font-semibold">{value}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-4 flex gap-x-2 justify-end">
          <Button variant="secondary" asChild>
            <Link to={edit}>Edit connection</Link>
          </Button>
          <Button asChild>
            <Link to={next}>Setup compete!</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function CompleteStep() {
  const { data: whoami } = useQuery(adminWhoami, {});

  return (
    <>
      <Card className="z-20 mt-16 w-full max-w-md mx-auto p-8">
        <div className="flex flex-col items-center justify-center gap-6">
          <div className="bg-green-200 rounded-full p-3">
            <CheckIcon className="w-8 h-8 text-green-500" />
          </div>
          <div className="text-center space-y-2">
            <h3 className="text-2xl font-bold">SAML Setup Complete</h3>
            <p className="text-muted-foreground">
              Congratulations! You have successfully set up your SAML
              connection.
            </p>
          </div>
          {whoami && (
            <Button className="w-full" asChild>
              {whoami.adminReturnUrl ? (
                <Link to={whoami.adminReturnUrl}>
                  {whoami.adminApplicationName
                    ? `Back to ${whoami.adminApplicationName}`
                    : "Exit this setup page"}
                </Link>
              ) : (
                <Link to="/">Exit this setup page</Link>
              )}
            </Button>
          )}
        </div>
      </Card>

      <a
        href="https://ssoready.com"
        className="fixed bottom-16 left-1/2 transform -translate-x-1/2 flex items-center gap-x-2 grayscale hover:grayscale-0"
      >
        <img src="/apple-touch-icon.png" alt="" className="h-4" />
        <span className="text-xs text-gray-500">Powered by SSOReady</span>
      </a>
    </>
  );
}
