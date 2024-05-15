import React, { useEffect, useState } from "react";
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

export function OnboardingPage() {
  const [searchParams] = useSearchParams();
  const onboardingInitialDemo =
    localStorage.getItem("onboarding_initial_demo") === "true";
  const hasSAMLAccessToken = !!searchParams.get("saml_access_code");
  console.log("initial step", hasSAMLAccessToken, onboardingInitialDemo);
  const [step, setStep] = useState(
    hasSAMLAccessToken && !onboardingInitialDemo ? 3 : 0,
  );
  const [apiKeySecretToken, setAPIKeySecretToken] = useState("");

  useEffect(() => {
    if (step === 0) {
      localStorage.setItem("onboarding_initial_demo", "true");
    }
    if (step === 2) {
      localStorage.removeItem("onboarding_initial_demo");
    }
  }, [step]);

  return (
    <div className="max-w-6xl mx-auto p-8 flex flex-col gap-y-8">
      <DemoCard
        done={step > 0}
        open={step === 0}
        onClickNext={() => setStep(1)}
        setAPIKeySecretToken={setAPIKeySecretToken}
      />
      <WhatJustHappenedCard
        done={step > 1}
        open={step === 1}
        onClickNext={() => setStep(2)}
      />
      <StartLoginCard
        done={step > 2}
        open={step === 2}
        apiKeySecretToken={apiKeySecretToken}
      />
      <HandleLoginCard
        done={false}
        open={step === 3}
        apiKeySecretToken={apiKeySecretToken}
      />
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

  const [searchParams, setSearchParams] = useSearchParams();
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

function WhatJustHappenedCard({
  open,
  done,
  onClickNext,
}: {
  open: boolean;
  done: boolean;
  onClickNext: () => void;
}) {
  const { data: onboardingState } = useQuery(getOnboardingState, {});

  return (
    onboardingState && (
      <Card className={clsx(done && "border-green-700")}>
        <CardHeader>
          <CardTitle>
            <div className="flex gap-x-4 items-center">
              <span>What just happened</span>
              {done && <CheckCircleIcon className="h-6 w-6 text-green-700" />}
            </div>
          </CardTitle>

          <CardDescription>
            What we did to make that demo work, and how you'll do it too.
          </CardDescription>
        </CardHeader>

        {open && (
          <CardContent className="text-sm">
            <div className="grid grid-cols-3 items-center gap-x-8">
              <div className="col-start-1 col-span-2">
                <p>The core thing you do with SSOReady is you:</p>
                <ol className="my-6 ml-6 list-decimal [&>li]:mt-2">
                  <li>
                    <span className="font-semibold">Start SAML Logins.</span>{" "}
                    You do that by asking us for a SAML redirect URL.
                  </li>
                  <li>
                    <span className="font-semibold">
                      Get details on SAML logins.
                    </span>{" "}
                    After your user does a SAML login, we redirect them back to
                    your app with a SAML access code. You exchange that code
                    with SSOReady for information about your user.
                  </li>
                </ol>
              </div>
              <div className="col-start-1 col-span-2">
                <p className="mb-6">
                  To make that happen, you need to set up three things in
                  SSOReady:
                </p>
              </div>
              <div className="col-start-1 col-span-2">
                <ol start={1} className="ml-6 list-decimal py-1">
                  <li>
                    An <span className="font-semibold">environment</span> tells
                    us where to redirect users back to.
                  </li>
                </ol>
              </div>
              <div className="col-span-1 py-1 text-muted-foreground text-xs">
                The demo-generated environment is{" "}
                <Link
                  className="underline underline-offset-2"
                  to={`/environments/${onboardingState.onboardingEnvironmentId}`}
                >
                  here
                </Link>
                .
              </div>
              <div className="col-start-1 col-span-2">
                <ol start={2} className="ml-6 list-decimal py-1">
                  <li>
                    Within an environment, an{" "}
                    <span className="font-semibold">organization</span> tells us
                    which tenant in your system a SAML login is for.
                  </li>
                </ol>
              </div>
              <div className="col-span-1 py-1 text-muted-foreground text-xs">
                The demo-generated organization is{" "}
                <Link
                  className="underline underline-offset-2"
                  to={`/environments/${onboardingState.onboardingEnvironmentId}/organizations/${onboardingState.onboardingOrganizationId}`}
                >
                  here
                </Link>
                .
              </div>
              <div className="col-start-1 col-span-2">
                <ol start={3} className="ml-6 list-decimal py-1">
                  <li>
                    Within an organization, a{" "}
                    <span className="font-semibold">SAML connection</span> tells
                    us about your customer's IDP, so we can carry out the SAML
                    flow.
                  </li>
                </ol>
              </div>
              <div className="col-span-1 py-1 text-muted-foreground text-xs">
                The demo-generated connection is{" "}
                <Link
                  className="underline underline-offset-2"
                  to={`/environments/${onboardingState.onboardingEnvironmentId}/organizations/${onboardingState.onboardingOrganizationId}/saml-connections/${onboardingState.onboardingSamlConnectionId}`}
                >
                  here
                </Link>
                .
              </div>
              <div className="col-start-1 col-span-2">
                <p className="mt-6">
                  You'll typically SAML connections and organizations whenever
                  you onboard a new enterprise-tier customer. You'll typically
                  create environments once, and then never touch them again.
                </p>
                <p className="mt-6">OK, enough concepts. Let's see the code.</p>
              </div>
            </div>

            <div className="mt-4 flex gap-x-2">
              <Button onClick={onClickNext}>
                Start a SAML login with SSOReady's SDK
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </div>
          </CardContent>
        )}
      </Card>
    )
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
  const code = `curl ${PUBLIC_API_URL}/v1/saml/redirect \\\n    -H "Authorization: Bearer ssoready_sk_•••••" \\\n    -d '{ "organization_id": "${onboardingState?.onboardingOrganizationId}" }'`;

  const [redirectURL, setRedirectURL] = useState("");
  const getSAMLRedirectURLMutation = useMutation(onboardingGetSAMLRedirectURL);

  return (
    onboardingState && (
      <Card className={clsx(done && "border-green-700")}>
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
              The demo generated an API key{" "}
              <span className="rounded border overflow-hidden inline-flex items-center">
                <code className="bg-black text-white px-1 py-0.5">
                  ssoready_sk_•••••
                </code>
                <span
                  className="p-1 cursor-pointer"
                  onClick={async () => {
                    await navigator.clipboard.writeText(apiKeySecretToken);
                  }}
                >
                  <CopyIcon className="h-4 w-4 text-muted-foreground" />
                </span>
              </span>{" "}
              scoped to a throwaway demo environment. Let's use it to request a
              "start a SAML login" redirect URL from the REST API.
            </p>

            <div className="rounded-lg overflow-hidden border">
              <div className="flex items-center">
                <span className="rounded px-2 py-1 m-2 bg-muted text-sm">
                  cURL
                </span>
              </div>
              <div className="w-full bg-black px-6 py-4 inline-block">
                <code className="text-sm text-white">
                  <pre>{code}</pre>
                </code>
              </div>
            </div>

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
                <div className="mt-4 rounded bg-black px-6 py-4 inline-block">
                  <code className="text-sm text-white">{redirectURL}</code>
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
  const code = `curl ${PUBLIC_API_URL}/v1/saml/redeem \\\n    -H "Authorization: Bearer ssoready_sk_•••••" \\\n    -d '{ "saml_access_token": "${samlAccessCode}" }'`;
  const { data: onboardingState } = useQuery(getOnboardingState, {});

  const [redeemData, setRedeemData] = useState<
    RedeemSAMLAccessCodeResponse | undefined
  >(undefined);
  const redeemSAMLAccessTokenMutation = useMutation(
    onboardingRedeemSAMLAccessToken,
  );

  return (
    onboardingState && (
      <Card className={clsx(done && "border-green-700")}>
        <CardHeader>
          <CardTitle>
            <div className="flex gap-x-4 items-center">
              <span>Handle a SAML Login</span>
              {done && <CheckCircleIcon className="h-6 w-6 text-green-700" />}
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
              with something we call a "SAML Access Token" in the query param
              (you can see it in the address bar right now).
            </p>

            <p className="mb-4 text-sm">
              You can use your API key again&mdash;
              <span className="rounded border overflow-hidden inline-flex items-center">
                <code className="bg-black text-white px-1 py-0.5">
                  ssoready_sk_•••••
                </code>
                <span
                  className="p-1 cursor-pointer"
                  onClick={async () => {
                    await navigator.clipboard.writeText(apiKeySecretToken);
                  }}
                >
                  <CopyIcon className="h-4 w-4 text-muted-foreground" />
                </span>
              </span>
              &mdash;to exchange that access token for details about the
              just-logged-in user.
            </p>

            <div className="rounded-lg overflow-hidden border">
              <div className="flex items-center">
                <span className="rounded px-2 py-1 m-2 bg-muted text-sm">
                  cURL
                </span>
              </div>
              <div className="w-full bg-black px-6 py-4 inline-block">
                <code className="text-sm text-white">
                  <pre>{code}</pre>
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
                <div className="mt-4 rounded bg-black px-6 py-4 block">
                  <code className="text-sm text-white">
                    <pre>{JSON.stringify(redeemData, null, 2)}</pre>
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
