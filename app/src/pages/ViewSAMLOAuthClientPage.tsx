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
  deleteSAMLOAuthClient,
  getSAMLOAuthClient,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { useNavigate, useParams } from "react-router";
import { Heading } from "lucide-react";
import { Button } from "@/components/ui/button";

export function ViewSAMLOAuthClientPage() {
  const { environmentId, samlOAuthClientId } = useParams();
  const { data: samlOAuthClient } = useQuery(getSAMLOAuthClient, {
    id: samlOAuthClientId,
  });

  const deleteSAMLOAuthClientMutation = useMutation(deleteSAMLOAuthClient);
  const navigate = useNavigate();
  const handleDelete = useCallback(async () => {
    await deleteSAMLOAuthClientMutation.mutateAsync({
      id: samlOAuthClientId,
    });
    navigate(`/environments/${environmentId}/saml-oauth-clients`);
  }, [
    samlOAuthClientId,
    environmentId,
    deleteSAMLOAuthClientMutation,
    navigate,
  ]);

  return (
    <div className="flex flex-col gap-y-8">
      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>SAML OAuth Client</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {samlOAuthClientId}
                </span>
              </div>

              <CardDescription>
                A SAML OAuth client is a way to add SAML support to an app that
                exclusively uses OAuth to do user logins.
              </CardDescription>
            </div>
          </div>
        </CardHeader>
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
                Revoke this SAML OAuth client and delete it. This cannot be
                undone.
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
