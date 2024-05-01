import React, { useCallback } from "react";
import { Button } from "@/components/ui/button";
import { ChevronLeft } from "lucide-react";
import { Link } from "react-router-dom";
import { useParams } from "react-router";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  getSAMLConnection,
  updateSAMLConnection,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { SAMLConnection } from "@/gen/ssoready/v1/ssoready_pb";

export function EditSAMLConnectionPage() {
  const { environmentId, organizationId, samlConnectionId } = useParams();
  const { data: samlConnection } = useQuery(getSAMLConnection, {
    id: samlConnectionId,
  });

  return (
    <>
      <div className="flex items-center gap-4">
        <Button variant="outline" size="icon" className="h-7 w-7">
          <ChevronLeft className="h-4 w-4" />
          <span className="sr-only">Back</span>
        </Button>
        <h1 className="flex-1 shrink-0 whitespace-nowrap text-xl font-semibold tracking-tight sm:grow-0">
          {samlConnectionId}
        </h1>
        {/*<Badge variant="outline" className="ml-auto sm:ml-0">*/}
        {/*  In stock*/}
        {/*</Badge>*/}
        <div className="hidden items-center gap-2 md:ml-auto md:flex">
          {/*<Button size="sm" asChild>*/}
          {/*  <Link*/}
          {/*      to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}/edit`}*/}
          {/*  >*/}
          {/*    Edit*/}
          {/*  </Link>*/}
          {/*</Button>*/}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Identity Provider (IdP) Settings</CardTitle>
          <CardDescription>
            These settings come from your customer's Identity Provider.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {samlConnection && (
            <UpdateSAMLConnectionForm samlConnection={samlConnection} />
          )}
        </CardContent>
      </Card>
    </>
  );
}

const FormSchema = z.object({
  idpEntityId: z.string().min(1, {
    message: "IDP Entity ID must be non-empty.",
  }),
  idpRedirectUrl: z.string().url({
    message: "IDP Redirect URL must be a valid URL.",
  }),
  idpCertificate: z.string().startsWith("-----BEGIN CERTIFICATE-----", {
    message: "IDP Certificate must be a PEM-encoded X.509 certificate.",
  }),
});

function UpdateSAMLConnectionForm({
  samlConnection,
}: {
  samlConnection: SAMLConnection;
}) {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      idpEntityId: samlConnection.idpEntityId,
      idpRedirectUrl: samlConnection.idpRedirectUrl,
      idpCertificate: samlConnection.idpCertificate,
    },
  });

  const updateSAMLConnectionMutation = useMutation(updateSAMLConnection);
  const handleSubmit = useCallback(
    async (data: z.infer<typeof FormSchema>) => {
      await updateSAMLConnectionMutation.mutateAsync({
        samlConnection: {
          id: samlConnection.id,
          idpEntityId: data.idpEntityId,
          idpRedirectUrl: data.idpRedirectUrl,
          idpCertificate: data.idpCertificate,
        },
      });
    },
    [updateSAMLConnection],
  );

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(handleSubmit)}
        className="w-2/3 space-y-6"
      >
        <FormField
          control={form.control}
          name="idpEntityId"
          render={({ field }) => (
            <FormItem>
              <FormLabel>IDP Entity ID</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormDescription>IDP Entity ID.</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="idpRedirectUrl"
          render={({ field }) => (
            <FormItem>
              <FormLabel>IDP Redirect URL</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormDescription>IDP Redirect URL.</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="idpCertificate"
          render={({ field }) => (
            <FormItem>
              <FormLabel>IDP Certificate</FormLabel>
              <FormControl>
                <Textarea {...field} />
              </FormControl>
              <FormDescription>
                IDP Certificate, as a PEM-encoded X.509 certificate. These start
                with '-----BEGIN CERTIFICATE-----' and end with '-----END
                CERTIFICATE-----'.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit">Submit</Button>
      </form>
    </Form>
  );
}
