import React, { useCallback, useEffect, useRef, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { v4 as uuidv4 } from "uuid";
import {
  API_URL,
  APP_URL,
  DUMMYIDP_CERTIFICATE,
  DUMMYIDP_URL,
  PUBLIC_API_URL,
} from "@/config";
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import {
  createAPIKey,
  createEnvironment,
  createOrganization,
  createSAMLConnection,
  getOnboardingState,
  getSAMLConnection,
  onboardingGetSAMLRedirectURL,
  onboardingRedeemSAMLAccessToken,
  redeemSAMLAccessCode,
  updateOnboardingState,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import axios from "axios";
import { RedeemSAMLAccessCodeResponse } from "@/gen/ssoready/v1/ssoready_pb";
import {
  ArrowRight,
  CheckCircleIcon,
  CheckIcon,
  CopyIcon,
  LoaderIcon,
  RefreshCcw,
} from "lucide-react";
import { clsx } from "clsx";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import hljs from "highlight.js/lib/core";
import formatXml from "xml-formatter";
import { offset, useFloating, useTransitionStyles } from "@floating-ui/react";
import { GoogleLogin } from "@react-oauth/google";

export function OnboardingPage() {
  const [searchParams] = useSearchParams();
  const onboardingInitialDemo =
    localStorage.getItem("onboarding_initial_demo") === "true";
  const hasSAMLAccessToken = !!searchParams.get("saml_access_code");
  const [step, setStep] = useState(
    hasSAMLAccessToken && !onboardingInitialDemo ? 3 : 0,
  );
  const [apiKeySecretToken, setAPIKeySecretToken] = useState("");

  useEffect(() => {
    if (step === 0) {
      localStorage.setItem("onboarding_initial_demo", "true");
    }
    if (step === 1) {
      localStorage.removeItem("onboarding_initial_demo");
    }
  }, [step]);

  return (
    <div className="max-w-6xl mx-auto p-8 flex flex-col gap-y-8">
      <DemoCard
        done={step > 0}
        open={step >= 0}
        onClickNext={() => setStep(1)}
        setAPIKeySecretToken={setAPIKeySecretToken}
      />
      {step >= 1 && (
        <StartLoginCard
          done={step > 1}
          open={step >= 1}
          apiKeySecretToken={apiKeySecretToken}
        />
      )}
      {step >= 2 && (
        <HandleLoginCard
          done={false}
          open={step >= 2}
          apiKeySecretToken={apiKeySecretToken}
        />
      )}
      <div className="h-[500px]" />
    </div>
  );
}

function DemoCard({
  done,
  open,
  onClickNext,
  setAPIKeySecretToken,
}: {
  done: boolean;
  open: boolean;
  onClickNext: () => void;
  setAPIKeySecretToken: (_: string) => void;
}) {
  const queryClient = useQueryClient();
  const { data: onboardingState } = useQuery(getOnboardingState, {});
  const getOnboardingStateMutation = useMutation(getOnboardingState);
  const getSAMLConnectionMutation = useMutation(getSAMLConnection);
  const createEnvironmentMutation = useMutation(createEnvironment);
  const createOrganizationMutation = useMutation(createOrganization);
  const createSAMLConnectionMutation = useMutation(createSAMLConnection);
  const createAPIKeyMutation = useMutation(createAPIKey);
  const updateOnboardingStateMutation = useMutation(updateOnboardingState);

  async function onClickLogin() {
    // upsert onboarding state if required
    if (onboardingState?.onboardingEnvironmentId === "") {
      const environment = await createEnvironmentMutation.mutateAsync({
        environment: {
          displayName: "Onboarding Demo",
          redirectUrl: `${APP_URL}/onboarding`,
        },
      });

      const organization = await createOrganizationMutation.mutateAsync({
        organization: {
          environmentId: environment.id,
          domains: ["yourcustomer.com"],
        },
      });

      const dummyIdpAppId = uuidv4();
      const samlConnection = await createSAMLConnectionMutation.mutateAsync({
        samlConnection: {
          organizationId: organization.id,
          idpEntityId: `https://dummyidp.com/apps/${dummyIdpAppId}`,
          idpRedirectUrl: `${DUMMYIDP_URL}/apps/${dummyIdpAppId}/sso`,
          idpCertificate: DUMMYIDP_CERTIFICATE,
          primary: true,
        },
      });

      await updateOnboardingStateMutation.mutateAsync({
        dummyidpAppId: dummyIdpAppId,
        onboardingEnvironmentId: environment.id,
        onboardingOrganizationId: organization.id,
        onboardingSamlConnectionId: samlConnection.id,
      });

      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey(getOnboardingState, {}),
      });
    }

    // hacky approach here to ensure we're reading the onboarding state written above
    const freshOnboardingState = await getOnboardingStateMutation.mutateAsync(
      {},
    );
    const samlConnection = await getSAMLConnectionMutation.mutateAsync({
      id: freshOnboardingState.onboardingSamlConnectionId,
    });

    const redirect = new URL(`${DUMMYIDP_URL}/instant-setup`);
    redirect.searchParams.set("appId", freshOnboardingState.dummyidpAppId);
    redirect.searchParams.set("spAcsUrl", samlConnection.spAcsUrl);
    redirect.searchParams.set("spEntityId", samlConnection.spEntityId);
    redirect.searchParams.set("email", "youruser@yourcustomer.com");
    redirect.searchParams.set("firstName", "Testy");
    redirect.searchParams.set("lastName", "McTestFace");

    window.location.href = redirect.toString();
  }

  const [searchParams] = useSearchParams();
  const samlAccessCode = searchParams.get("saml_access_code");
  const onboardingRedeeemSAMLAccessToken = useMutation(
    onboardingRedeemSAMLAccessToken,
  );
  const [redeemResponse, setRedeemResponse] = useState<
    RedeemSAMLAccessCodeResponse | undefined
  >(undefined);
  useEffect(() => {
    if (!samlAccessCode || !onboardingState) {
      return;
    }

    (async () => {
      const apiKey = await createAPIKeyMutation.mutateAsync({
        apiKey: {
          environmentId: onboardingState.onboardingEnvironmentId,
        },
      });

      const redeemResponse = await onboardingRedeeemSAMLAccessToken.mutateAsync(
        {
          accessCode: samlAccessCode,
          apiKeySecretToken: apiKey.secretToken,
        },
      );

      setAPIKeySecretToken(apiKey.secretToken);
      setRedeemResponse(redeemResponse);
    })();
  }, [samlAccessCode, onboardingState, createAPIKeyMutation.mutateAsync]);

  return (
    <Card className={clsx(done && "border-green-700")}>
      <CardHeader>
        <CardTitle>
          <div className="flex gap-x-4 items-center">
            <span>Sign in using SSOReady</span>
            {done && <CheckCircleIcon className="h-6 w-6 text-green-700" />}
          </div>
        </CardTitle>
        <CardDescription>
          Let's start with a demo of what your users will see.
        </CardDescription>
      </CardHeader>

      {open && (
        <CardContent>
          <DemoLogin
            onClickLogin={onClickLogin}
            redeemResponse={redeemResponse}
          />

          {redeemResponse && (
            <div className="mt-4 flex justify-end gap-x-2">
              <Button
                variant="secondary"
                onClick={() => {
                  setRedeemResponse(undefined);
                }}
              >
                <RefreshCcw className="mr-2 h-4 w-4" />
                Run the demo again
              </Button>
              <Button onClick={onClickNext}>
                See how you'd do this yourself
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </div>
          )}
        </CardContent>
      )}
    </Card>
  );
}

function DemoLogin({
  onClickLogin,
  redeemResponse,
}: {
  onClickLogin: () => void;
  redeemResponse?: RedeemSAMLAccessCodeResponse;
}) {
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const timeoutID = setTimeout(() => {
      setLoading(false);
    }, 250);

    return () => clearTimeout(timeoutID);
  });

  return (
    <div className="border rounded-lg overflow-hidden">
      <div className="px-12 py-16 bg-gradient-to-b from-[#49CBF3] to-[#F0361C]">
        <div className="rounded-lg overflow-hidden shadow-lg">
          <header className="flex items-center h-12 px-4 border-b bg-gray-100 dark:bg-gray-900 dark:border-gray-800">
            <div className="relative">
              <div className="absolute left-2 top-1/2 -translate-y-1/2 flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-red-500" />
                <div className="w-3 h-3 rounded-full bg-yellow-500" />
                <div className="w-3 h-3 rounded-full bg-green-500" />
              </div>
            </div>

            <div className="flex-1 max-w-md mx-auto">
              <div className="relative">
                <div className="pl-4 w-full h-8 bg-white rounded-md border border-gray-200 text-sm flex items-center text-muted-foreground">
                  https://
                  <span className="text-foreground">yourcompany.com</span>
                  /sign-in
                </div>
              </div>
            </div>
          </header>

          <div className="dark bg-muted p-8 min-h-[400px] flex items-center justify-center">
            <Card className="max-w-sm">
              {redeemResponse && loading && (
                <>
                  <CardHeader className="flex flex-col items-center space-y-2">
                    <LoaderIcon className="h-12 w-12 animate-spin" />
                  </CardHeader>
                </>
              )}

              {redeemResponse && !loading && (
                <>
                  <CardHeader className="flex flex-col items-center space-y-2">
                    <CheckIcon className="h-12 w-12 text-green-500" />
                    <CardTitle className="text-2xl font-bold">
                      Hi, {redeemResponse.attributes["firstName"]}{" "}
                      {redeemResponse.attributes["lastName"]}!
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    Your email is: {redeemResponse.email}
                  </CardContent>
                </>
              )}
              {!redeemResponse && (
                <>
                  <CardHeader>
                    <CardTitle className="text-2xl">Login</CardTitle>
                    <CardDescription>
                      Enter your email below to login to your account
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid gap-4">
                      <div className="grid gap-2">
                        <Label htmlFor="email">Email</Label>
                        <Input
                          id="email"
                          type="email"
                          value="youruser@yourcustomer.com"
                          required
                          disabled
                        />
                      </div>
                      <Button
                        type="submit"
                        className="w-full"
                        onClick={onClickLogin}
                      >
                        Log in with SSO
                      </Button>
                    </div>
                  </CardContent>
                </>
              )}
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}

function StartLoginCard({
  open,
  done,
  apiKeySecretToken,
}: {
  open: boolean;
  done: boolean;
  apiKeySecretToken: string;
}) {
  const { data: onboardingState } = useQuery(getOnboardingState, {});
  const code = `curl ${PUBLIC_API_URL}/v1/saml/redirect \\\n    -H "Content-Type: application/json" \\\n    -H "Authorization: Bearer ssoready_sk_•••••" \\\n    -d '{ "organization_id": "${onboardingState?.onboardingOrganizationId}" }'`;
  const copyCode = `curl ${PUBLIC_API_URL}/v1/saml/redirect \\\n    -H "Content-Type: application/json" \\\n    -H "Authorization: Bearer ${apiKeySecretToken}" \\\n    -d '{ "organization_id": "${onboardingState?.onboardingOrganizationId}" }'`;

  const [redirectURL, setRedirectURL] = useState("");
  const getSAMLRedirectURLMutation = useMutation(onboardingGetSAMLRedirectURL);

  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (ref.current && !done) {
      console.log("scroll to StartLoginCard");
      ref.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [done]);

  return (
    onboardingState && (
      <Card className={clsx(done && "border-green-700")} ref={ref}>
        <CardHeader>
          <CardTitle>
            <div className="flex gap-x-4 items-center">
              <span>Start a SAML Login</span>
              {done && <CheckCircleIcon className="h-6 w-6 text-green-700" />}
            </div>
          </CardTitle>

          <CardDescription>
            Use our REST API to get a "log in with SSO" redirect URL.
          </CardDescription>
        </CardHeader>

        {open && (
          <CardContent>
            <p className="mb-4 text-sm">
              In the demo above, the "Log in with SSO" button redirected your
              user to their company's Identity Provider. You get that URL by
              calling SSOReady's "get SAML redirect URL" endpoint.
            </p>

            <div className="rounded-lg overflow-hidden border">
              <div className="m-2 flex items-center justify-between">
                <div>
                  <span className="rounded px-2 py-1 bg-muted text-sm">
                    cURL
                  </span>
                </div>
                <CopyButton copyText={copyCode} />
              </div>
              <div className="text-xs font-mono bg-gray-100 py-2 px-4 max-w-full overflow-auto">
                <code>
                  <pre
                    dangerouslySetInnerHTML={{
                      __html: hljs.highlight(code, {
                        language: "bash",
                      }).value,
                    }}
                  />
                </code>
              </div>
            </div>

            <p className="mt-4 text-sm">
              (How you get an organization ID is covered in our docs.)
            </p>

            <Button
              disabled={redirectURL !== ""}
              className="mt-4"
              onClick={async () => {
                const { redirectUrl } =
                  await getSAMLRedirectURLMutation.mutateAsync({
                    apiKeySecretToken,
                    samlConnectionId:
                      onboardingState.onboardingSamlConnectionId,
                  });

                setRedirectURL(redirectUrl);
              }}
            >
              Generate SAML Redirect URL
              {redirectURL !== "" && <CheckIcon className="ml-2 h-4 w-4" />}
            </Button>

            {redirectURL !== "" && (
              <>
                <div className="mt-4 rounded text-xs font-mono bg-gray-100 py-2 px-4 max-w-full overflow-auto">
                  <code>
                    <pre
                      dangerouslySetInnerHTML={{
                        __html: hljs.highlight(
                          JSON.stringify({ redirectUrl: redirectURL }, null, 4),
                          {
                            language: "json",
                          },
                        ).value,
                      }}
                    />
                  </code>
                </div>

                <div className="mt-4 flex gap-x-4 items-center">
                  <Button asChild>
                    <Link to={redirectURL}>Follow SAML redirect URL</Link>
                  </Button>

                  <span className="text-sm text-muted-foreground">
                    This link will take you to DummyIDP. Once you "log in"
                    you'll be redirected back here.
                  </span>
                </div>
              </>
            )}
          </CardContent>
        )}
      </Card>
    )
  );
}

function HandleLoginCard({
  open,
  done,
  apiKeySecretToken,
}: {
  open: boolean;
  done: boolean;
  apiKeySecretToken: string;
}) {
  const [searchParams] = useSearchParams();
  const samlAccessCode = searchParams.get("saml_access_code");
  const code = `curl ${PUBLIC_API_URL}/v1/saml/redeem \\\n    -H "Content-Type: application/json" \\\n    -H "Authorization: Bearer ssoready_sk_•••••" \\\n    -d '{ "access_code": "${samlAccessCode}" }'`;
  const copyCode = `curl ${PUBLIC_API_URL}/v1/saml/redeem \\\n    -H "Content-Type: application/json" \\\n    -H "Authorization: Bearer ${apiKeySecretToken}" \\\n    -d '{ "access_code": "${samlAccessCode}" }'`;
  const { data: onboardingState } = useQuery(getOnboardingState, {});

  const [redeemData, setRedeemData] = useState<
    RedeemSAMLAccessCodeResponse | undefined
  >(undefined);
  const redeemSAMLAccessTokenMutation = useMutation(
    onboardingRedeemSAMLAccessToken,
  );

  const scrollTo = useCallback((node: any) => {
    if (node !== null) {
      node.scrollIntoView({ behavior: "smooth" });
    }
  }, []);

  return (
    onboardingState && (
      <Card className={clsx(redeemData && "border-green-700")} ref={scrollTo}>
        <CardHeader>
          <CardTitle>
            <div className="flex gap-x-4 items-center">
              <span>Handle a SAML Login</span>
              {redeemData && (
                <CheckCircleIcon className="h-6 w-6 text-green-700" />
              )}
            </div>
          </CardTitle>

          <CardDescription>
            Use our REST API to get details about who just logged in via SAML.
          </CardDescription>
        </CardHeader>

        {open && (
          <CardContent>
            <p className="mb-4 text-sm">
              Once a SAML login happens, we redirect your user back to your app,
              with something we call a "SAML Access Code" in the query param
              (you can see it in the address bar right now).
            </p>

            <p className="mb-4 text-sm">
              You can use your API key again to exchange that access code for
              details about the just-logged-in user.
            </p>

            <div className="rounded-lg overflow-hidden border">
              <div className="m-2 flex items-center justify-between">
                <div>
                  <span className="rounded px-2 py-1 bg-muted text-sm">
                    cURL
                  </span>
                </div>
                <CopyButton copyText={copyCode} />
              </div>
              <div className="text-xs font-mono bg-gray-100 py-2 px-4 max-w-full overflow-auto">
                <code>
                  <pre
                    dangerouslySetInnerHTML={{
                      __html: hljs.highlight(code, {
                        language: "bash",
                      }).value,
                    }}
                  />
                </code>
              </div>
            </div>

            <Button
              disabled={!!redeemData}
              className="mt-4"
              onClick={async () => {
                const redeemData =
                  await redeemSAMLAccessTokenMutation.mutateAsync({
                    apiKeySecretToken,
                    accessCode: samlAccessCode!,
                  });

                setRedeemData(redeemData);
              }}
            >
              Redeem SAML Access Code
              {!!redeemData && <CheckIcon className="ml-2 h-4 w-4" />}
            </Button>

            {redeemData !== undefined && (
              <>
                <div className="mt-4 rounded text-xs font-mono bg-gray-100 py-2 px-4 max-w-full overflow-auto">
                  <code>
                    <pre
                      dangerouslySetInnerHTML={{
                        __html: hljs.highlight(
                          JSON.stringify(
                            {
                              email: redeemData.email,
                              attributes: redeemData.attributes,
                              organizationId: redeemData.organizationId,
                              samlFlowId: redeemData.samlFlowId,
                            },
                            null,
                            4,
                          ),
                          {
                            language: "json",
                          },
                        ).value,
                      }}
                    />
                  </code>
                </div>

                <p className="mt-4 text-sm">
                  From here, it's your job to create a session for{" "}
                  <code className="rounded bg-black text-white px-1 py-0.5">
                    {redeemData.email}
                  </code>
                  . That's all there is to it!
                </p>

                <div className="mt-4 flex justify-end gap-x-2">
                  <Button variant="secondary" asChild>
                    <Link to="/">Exit this onboarding flow</Link>
                  </Button>
                  <Button asChild>
                    <a href="https://ssoready.com/docs">Check out the docs</a>
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        )}
      </Card>
    )
  );
}

function CopyButton({ copyText }: { copyText: string }) {
  const [open, setOpen] = useState(false);
  const { refs, floatingStyles, context } = useFloating({
    open,
    onOpenChange: setOpen,
    placement: "top",
    middleware: [offset(5)],
  });
  const { isMounted, styles } = useTransitionStyles(context);

  useEffect(() => {
    if (open) {
      const timeoutId = setTimeout(() => {
        setOpen(false);
      }, 1000);
      return () => clearTimeout(timeoutId);
    }
  }, [open]);

  return (
    <div>
      <CopyIcon
        ref={refs.setReference}
        className="h-4 w-4 cursor-pointer text-muted-foreground"
        onClick={async () => {
          await navigator.clipboard.writeText(copyText);
          setOpen(true);
        }}
      />
      {open && (
        <div ref={refs.setFloating} style={floatingStyles}>
          {isMounted && (
            <div
              style={styles}
              className="bg-black text-white p-1 text-xs rounded"
            >
              Copied!
            </div>
          )}
        </div>
      )}
    </div>
  );
}
