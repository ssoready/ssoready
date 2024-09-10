import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import React, { useCallback } from "react";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  deleteAPIKey,
  getAPIKey,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { useNavigate, useParams } from "react-router";
import { Heading } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Title } from "@/components/Title";
import { InfoTooltip } from "@/components/InfoTooltip";
import { DocsLink } from "@/components/DocsLink";

export function ViewAPIKeyPage() {
  const { environmentId, apiKeyId } = useParams();
  const { data: apiKey } = useQuery(getAPIKey, { id: apiKeyId });

  const deleteAPIKeyMutation = useMutation(deleteAPIKey);
  const navigate = useNavigate();
  const handleDelete = useCallback(async () => {
    await deleteAPIKeyMutation.mutateAsync({
      id: apiKeyId,
    });
    navigate(`/environments/${environmentId}`);
  }, [apiKeyId, environmentId, deleteAPIKeyMutation, navigate]);

  return (
    <div className="flex flex-col gap-y-8">
      <Title title="API Key" />

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>API Key</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {apiKeyId}
                </span>
              </div>

              <CardDescription>
                An API key is how your application authenticates with SSOReady's
                SDKs or REST APIs.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/environments#api-keys" />
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-5 gap-y-2">
            <div className="text-sm col-span-2 text-muted-foreground flex items-center gap-x-2">
              Management API Access
              <InfoTooltip>
                Whether this API key can use the Management API.
                <DocsLink to="https://ssoready.com/docs/ssoready-concepts/environments#management-api" />
              </InfoTooltip>
            </div>
            <div className="text-sm col-span-3">
              {apiKey?.hasManagementApiAccess ? "Yes" : "No"}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className="border-destructive">
        <CardHeader>
          <CardTitle>Danger Zone</CardTitle>
        </CardHeader>

        <CardContent>
          <div className="flex justify-between items-center">
            <div>
              <div className="text-sm font-semibold">Revoke and Delete</div>
              <p className="text-sm">
                Revoke this API key and delete it. This cannot be undone.
              </p>
            </div>

            <Button variant="destructive" onClick={handleDelete}>
              Revoke & Delete
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
