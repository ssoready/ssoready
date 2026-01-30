import { LayoutMain } from "@/components/Layout";
import { ValueCopier } from "@/components/ValueCopier";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Switch } from "@/components/ui/switch";
import {
  adminGetSCIMDirectory,
  adminRotateSCIMDirectoryBearerToken,
  adminUpdateSCIMDirectory,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { SCIMDirectory } from "@/gen/ssoready/v1/ssoready_pb";
import { useTitle } from "@/useTitle";
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from "@connectrpc/connect-query";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { CircleAlert } from "lucide-react";
import React, { useCallback, useState } from "react";
import { Helmet } from "react-helmet";
import { useForm } from "react-hook-form";
import { useParams } from "react-router";
import { z } from "zod";

export function ViewSCIMDirectoryPage() {
  const { scimDirectoryId } = useParams();
  const title = useTitle(`SCIM Directory ${scimDirectoryId}`);

  const { data: scimDirectory } = useQuery(adminGetSCIMDirectory, {
    id: scimDirectoryId,
  });

  const [warnBearerTokenAlertOpen, setWarnBearerTokenAlertOpen] =
    useState(false);

  const [bearerTokenAlertOpen, setBearerTokenAlertOpen] = useState(false);
  const [bearerToken, setBearerToken] = useState("");
  const rotateSCIMDirectoryBearerTokenMutation = useMutation(
    adminRotateSCIMDirectoryBearerToken,
  );
  const queryClient = useQueryClient();
  const handleRotateBearerToken = async () => {
    const { bearerToken } =
      await rotateSCIMDirectoryBearerTokenMutation.mutateAsync({
        scimDirectoryId,
      });
    setBearerToken(bearerToken);
    setBearerTokenAlertOpen(true);

    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey(adminGetSCIMDirectory, {
        id: scimDirectoryId,
      }),
    });
  };

  return (
    <LayoutMain>
      <Helmet>
        <title>{title}</title>
      </Helmet>
      <AlertDialog
        open={warnBearerTokenAlertOpen}
        onOpenChange={setWarnBearerTokenAlertOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
              Rotating a SCIM directory's bearer token will break any existing
              syncs until you input the new bearer token into your identity
              provider settings.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleRotateBearerToken}>
              I understand, rotate bearer token
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog
        open={bearerTokenAlertOpen}
        onOpenChange={setBearerTokenAlertOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Updated SCIM Bearer Token</AlertDialogTitle>
            <AlertDialogDescription>
              This SCIM directory's bearer token has been updated.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <Alert>
            <CircleAlert className="h-4 w-4" />
            <AlertTitle>Copy this bearer token</AlertTitle>
            <AlertDescription>
              Input this authentication bearer token into your identity provider
              settings.
            </AlertDescription>
          </Alert>

          <div className="text-sm font-medium leading-none">
            SCIM Bearer Token
          </div>

          <ValueCopier value={bearerToken} />

          <AlertDialogFooter>
            <AlertDialogCancel>Close</AlertDialogCancel>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <div className="flex flex-col gap-y-8">
        <Card>
          <CardHeader>
            <div className="flex justify-between items-center">
              <div className="flex flex-col space-y-1.5">
                <div className="flex gap-4">
                  <CardTitle>SCIM Directory</CardTitle>

                  <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                    {scimDirectoryId}
                  </span>
                </div>
              </div>

              <div>
                {scimDirectory?.scimDirectory && (
                  <EditSCIMDirectoryAlertDialog
                    scimDirectory={scimDirectory.scimDirectory}
                  />
                )}
              </div>
            </div>
          </CardHeader>

          <CardContent>
            <div className="grid grid-cols-4 gap-y-2">
              <div className="text-sm col-span-1 text-muted-foreground">
                Primary
              </div>
              <div className="text-sm col-span-3">
                {scimDirectory?.scimDirectory?.primary ? "Yes" : "No"}
              </div>
              <div className="text-sm col-span-1 text-muted-foreground">
                SCIM Base URL
              </div>
              <div className="col-span-3">
                {scimDirectory?.scimDirectory?.scimBaseUrl && (
                  <ValueCopier
                    value={scimDirectory.scimDirectory.scimBaseUrl}
                  />
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>SCIM Authentication</CardTitle>
            <CardDescription>
              To configure a SCIM directory with our product, you need to input
              a authentication bearer token into your identity provider. You can
              only view a bearer token immediately after creating it; we only
              store a hash.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {scimDirectory?.scimDirectory?.hasClientBearerToken ? (
              <Button
                variant="outline"
                onClick={() => setWarnBearerTokenAlertOpen(true)}
              >
                Rotate bearer token
              </Button>
            ) : (
              <Button variant="outline" onClick={handleRotateBearerToken}>
                Generate bearer token
              </Button>
            )}
          </CardContent>
        </Card>
      </div>
    </LayoutMain>
  );
}

const FormSchema = z.object({
  primary: z.boolean(),
});

function EditSCIMDirectoryAlertDialog({
  scimDirectory,
}: {
  scimDirectory: SCIMDirectory;
}) {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      primary: scimDirectory.primary,
    },
  });

  const [open, setOpen] = useState(false);
  const updateSCIMDirectoryMutation = useMutation(adminUpdateSCIMDirectory);
  const queryClient = useQueryClient();
  const handleSubmit = useCallback(
    async (values: z.infer<typeof FormSchema>, e: any) => {
      e.preventDefault();
      await updateSCIMDirectoryMutation.mutateAsync({
        scimDirectory: {
          id: scimDirectory.id,
          primary: values.primary,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey(adminGetSCIMDirectory, {
          id: scimDirectory.id,
        }),
      });

      setOpen(false);
    },
    [setOpen, scimDirectory, updateSCIMDirectoryMutation, queryClient],
  );

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        <Button variant="outline">Edit</Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)}>
            <AlertDialogHeader>
              <AlertDialogTitle>Edit SCIM directory</AlertDialogTitle>
            </AlertDialogHeader>

            <div className="my-4 space-y-4">
              <FormField
                control={form.control}
                name="primary"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Primary</FormLabel>
                    <FormControl className="block">
                      <Switch
                        name={field.name}
                        id={field.name}
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                    <FormDescription>
                      Whether this is the preferred SCIM directory to use by
                      default.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <Button type="submit">Save</Button>
            </AlertDialogFooter>
          </form>
        </Form>
      </AlertDialogContent>
    </AlertDialog>
  );
}
