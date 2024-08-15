import { z } from "zod";
import { Environment } from "@/gen/ssoready/v1/ssoready_pb";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import React, { useCallback, useState } from "react";
import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import {
  createEnvironment,
  getEnvironment,
  listEnvironments,
  updateEnvironment,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useNavigate } from "react-router";

const FormSchema = z.object({
  displayName: z.string().min(1, {
    message: "Display name is required.",
  }),
  redirectUrl: z
    .string()
    .url({
      message: "Redirect URL must be a valid URL.",
    })
    .refine((arg) => !arg.includes(" "), {
      message: "Redirect URL must be a valid URL.",
    }),
  authUrl: z.string(),
});

export function CreateEnvironmentPage() {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      displayName: "",
      redirectUrl: "",
      authUrl: "",
    },
  });

  const createEnvironmentMutation = useMutation(createEnvironment);
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const handleSubmit = useCallback(
    async (values: z.infer<typeof FormSchema>, e: any) => {
      e.preventDefault();
      const environment = await createEnvironmentMutation.mutateAsync({
        environment: {
          displayName: values.displayName,
          redirectUrl: values.redirectUrl,
          authUrl: values.authUrl,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey(listEnvironments, {}),
      });

      navigate(`/environments/${environment.id}`);
    },
    [createEnvironmentMutation, queryClient, navigate],
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create environment</CardTitle>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)}>
            <div className="space-y-4">
              <FormField
                control={form.control}
                name="displayName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Display name</FormLabel>
                    <FormControl>
                      <Input placeholder="Production" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="redirectUrl"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Redirect URL</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="http://localhost:8080/ssoready/callback"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      After a SAML login, your users get redirected to this
                      address. You usually want to point this at an
                      SSOReady-specific page on your web application.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <Button className="mt-8" type="submit">
              Save
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
