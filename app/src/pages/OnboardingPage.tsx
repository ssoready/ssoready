import React from "react";
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
import { useNavigate } from "react-router";
import { v4 as uuidv4 } from "uuid";
import { APP_URL, DUMMYIDP_CERTIFICATE, DUMMYIDP_URL } from "@/config";
import { useMutation } from "@connectrpc/connect-query";
import {
  createEnvironment,
  createOrganization,
  createSAMLConnection,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";

export function OnboardingPage() {
  const navigate = useNavigate();
  const createEnvironmentMutation = useMutation(createEnvironment);
  const createOrganizationMutation = useMutation(createOrganization);
  const createSAMLConnectionMutation = useMutation(createSAMLConnection);

  async function onClickLogin() {
    const environment = await createEnvironmentMutation.mutateAsync({
      environment: {
        displayName: "Onboarding Demo",
        redirectUrl: `${APP_URL}/onboarding/callback`,
      },
    });

    const organization = await createOrganizationMutation.mutateAsync({
      organization: {
        environmentId: environment.id,
        domains: ["yourcustomer.com"],
        externalId: "YOUR_DATABASE_ID_FOR_THE_CUSTOMER",
      },
    });

    const dummyIdpAppId = uuidv4();
    const samlConnection = await createSAMLConnectionMutation.mutateAsync({
      samlConnection: {
        organizationId: organization.id,
        idpEntityId: `https://dummyidp.com/apps/${dummyIdpAppId}`,
        idpRedirectUrl: `${DUMMYIDP_URL}/apps/${dummyIdpAppId}/sso`,
        idpCertificate: DUMMYIDP_CERTIFICATE,
      },
    });

    console.log(JSON.stringify(environment));
    console.log(JSON.stringify(organization));
    console.log(JSON.stringify(samlConnection));
    console.log(dummyIdpAppId);

    const redirect = new URL(`${DUMMYIDP_URL}/instant-setup`);
    redirect.searchParams.set("appId", dummyIdpAppId);
    redirect.searchParams.set("spAcsUrl", samlConnection.spAcsUrl);
    redirect.searchParams.set("spEntityId", samlConnection.spEntityId);
    redirect.searchParams.set("email", "youruser@yourcustomer.com");
    redirect.searchParams.set("firstName", "Testy");
    redirect.searchParams.set("lastName", "McTestFace");

    window.location.href = redirect.toString();
  }

  return (
    <div className="max-w-6xl mx-auto p-8 flex flex-col gap-y-8">
      <Card>
        <CardHeader>
          <CardTitle>Sign in using SSOReady</CardTitle>
          <CardDescription>
            Let's start with a demo of what your users will see.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <DemoLogin onClickLogin={onClickLogin} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Create an environment</CardTitle>
        </CardHeader>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Create an organization and SAML connection</CardTitle>
        </CardHeader>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Get a Redirect URL</CardTitle>
        </CardHeader>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Redeem the SAML Access Token</CardTitle>
        </CardHeader>
      </Card>
    </div>
  );
}

function DemoLogin({ onClickLogin }: { onClickLogin: () => void }) {
  return (
    <div className="dark bg-background p-8 rounded">
      <Card className="mx-auto max-w-sm">
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
            <Button type="submit" className="w-full" onClick={onClickLogin}>
              Log in with SSO
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
