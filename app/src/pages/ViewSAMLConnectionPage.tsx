import React, { useCallback, useId, useState } from "react";
import { useMatch, useParams } from "react-router";
import { useInfiniteQuery, useQuery } from "@connectrpc/connect-query";
import {
  getEnvironment,
  getOrganization,
  getSAMLConnection,
  listOrganizations,
  listSAMLConnections,
  listSAMLFlows,
  parseSAMLMetadata,
  updateSAMLConnection,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ChevronLeft } from "lucide-react";
import { Link } from "react-router-dom";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import moment from "moment";
import { z } from "zod";
import { SAMLConnection, SAMLFlowStatus } from "@/gen/ssoready/v1/ssoready_pb";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
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
import { Textarea } from "@/components/ui/textarea";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { useQueryClient } from "@tanstack/react-query";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Label } from "@/components/ui/label";

export function ViewSAMLConnectionPage() {
  const { environmentId, organizationId, samlConnectionId } = useParams();
  const { data: samlConnection } = useQuery(getSAMLConnection, {
    id: samlConnectionId,
  });
  const flowsPathMatch = useMatch(
    "/environments/:environmentId/organizations/:organizationId/saml-connections/:samlConnectionId/flows",
  );

  return (
    <div className="flex flex-col gap-y-8">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link
                to={`/environments/${environmentId}/organizations/${organizationId}`}
              >
                {organizationId}
              </Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{samlConnectionId}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div className="flex flex-col space-y-1.5">
              <div className="flex gap-4">
                <CardTitle>SAML Connection</CardTitle>

                <span className="text-xs font-mono bg-gray-100 py-1 px-2 rounded-sm">
                  {samlConnectionId}
                </span>
              </div>

              <CardDescription>
                A SAML connection is a link between your product and your
                customer's Identity Provider.
              </CardDescription>
            </div>
          </div>
        </CardHeader>
      </Card>

      <Tabs defaultValue={flowsPathMatch ? "flows" : "config"}>
        <TabsList>
          <TabsTrigger value="config" asChild>
            <Link
              to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}`}
            >
              Configuration
            </Link>
          </TabsTrigger>
          <TabsTrigger value="flows" asChild>
            <Link
              to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}/flows`}
            >
              Login Flows
            </Link>
          </TabsTrigger>
        </TabsList>
        <TabsContent value="config">
          <Card>
            <CardHeader>
              <CardTitle>Service Provider Configuration</CardTitle>
              <CardDescription>
                The configuration here is assigned automatically by SSOReady,
                and needs to be inputted into your customer's Identity Provider
                by their IT admin.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-4 gap-y-2 items-center">
                <div className="text-sm col-span-1 text-muted-foreground">
                  Entity ID
                </div>
                <div className="text-sm col-span-3">
                  {samlConnection?.spEntityId}
                </div>

                <div className="text-sm col-span-1 text-muted-foreground">
                  Assertion Consumer Service (ACS) URL
                </div>
                <div className="text-sm col-span-3">
                  {samlConnection?.spAcsUrl}
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="mt-4">
            <CardHeader>
              <div className="flex justify-between items-center">
                <div className="flex flex-col space-y-1.5">
                  <CardTitle>Identity Provider Configuration</CardTitle>
                  <CardDescription>
                    The configuration here needs to be copied over from the
                    customer's Identity Provider.
                  </CardDescription>
                </div>

                {samlConnection && (
                  <EditSAMLConnectionAlertDialog
                    samlConnection={samlConnection}
                  />
                )}
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-4 gap-y-2 items-center">
                <div className="text-sm col-span-1 text-muted-foreground">
                  Entity ID
                </div>
                <div className="text-sm col-span-3">
                  {samlConnection?.idpEntityId}
                </div>
                <div className="text-sm col-span-1 text-muted-foreground">
                  Redirect URL
                </div>
                <div className="text-sm col-span-3">
                  {samlConnection?.idpRedirectUrl}
                </div>
              </div>

              <Collapsible className="mt-1.5">
                <CollapsibleTrigger className="text-sm text-muted-foreground">
                  Certificate (click to show)
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <div className="bg-black rounded-lg px-6 py-4 mt-4 inline-block">
                    <code className="text-sm text-white">
                      <pre>{samlConnection?.idpCertificate}</pre>
                    </code>
                  </div>
                </CollapsibleContent>
              </Collapsible>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="flows">
          <ListLoginFlowsTabContent />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function ListLoginFlowsTabContent() {
  const { environmentId, organizationId, samlConnectionId } = useParams();
  const {
    data: listFlowsResponses,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery(
    listSAMLFlows,
    { samlConnectionId, pageToken: "" },
    {
      pageParamKey: "pageToken",
      getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
    },
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>SAML Login Flows</CardTitle>
        <CardDescription>
          SAML login flows from this connection are listed here.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>Timestamp</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>User ID</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {listFlowsResponses?.pages
              ?.flatMap((page) => page.samlFlows)
              ?.map((flow) => (
                <TableRow key={flow.id}>
                  <TableCell className="max-w-[200px] truncate">
                    <Link
                      to={`/environments/${environmentId}/organizations/${organizationId}/saml-connections/${samlConnectionId}/flows/${flow.id}`}
                      className="underline underline-offset-4 decoration-muted-foreground"
                    >
                      {flow.id}
                    </Link>
                  </TableCell>
                  <TableCell>
                    {moment(flow.createTime!.toDate()).format()}
                  </TableCell>
                  <TableCell>
                    {flow.status ===
                      SAMLFlowStatus.SAML_FLOW_STATUS_IN_PROGRESS && (
                      <Badge variant="secondary">In progress</Badge>
                    )}
                    {flow.status === SAMLFlowStatus.SAML_FLOW_STATUS_FAILED && (
                      <Badge variant="destructive">Failed</Badge>
                    )}
                    {flow.status ===
                      SAMLFlowStatus.SAML_FLOW_STATUS_SUCCEEDED && (
                      <Badge>Succeeded</Badge>
                    )}
                  </TableCell>
                  <TableCell>{flow.subjectIdpId}</TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>

        {hasNextPage && (
          <Button variant="secondary" onClick={() => fetchNextPage()}>
            Load more
          </Button>
        )}
      </CardContent>
    </Card>
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

function EditSAMLConnectionAlertDialog({
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

  const [open, setOpen] = useState(false);
  const updateSAMLConnectionMutation = useMutation(updateSAMLConnection);
  const queryClient = useQueryClient();

  const handleSubmit = useCallback(
    async (data: z.infer<typeof FormSchema>, e: any) => {
      e.preventDefault();
      await updateSAMLConnectionMutation.mutateAsync({
        samlConnection: {
          id: samlConnection.id,
          idpEntityId: data.idpEntityId,
          idpRedirectUrl: data.idpRedirectUrl,
          idpCertificate: data.idpCertificate,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey(getSAMLConnection, {
          id: samlConnection.id,
        }),
      });

      setOpen(false);
    },
    [samlConnection.id, updateSAMLConnectionMutation, queryClient, setOpen],
  );

  const id = useId();
  const [metadataUrl, setMetadataUrl] = useState("");
  const parseSAMLMetadataMutation = useMutation(parseSAMLMetadata);
  const handleLoadMetadata = useCallback(async () => {
    const { idpRedirectUrl, idpCertificate, idpEntityId } =
      await parseSAMLMetadataMutation.mutateAsync({ url: metadataUrl });

    form.setValue("idpRedirectUrl", idpRedirectUrl);
    form.setValue("idpCertificate", idpCertificate);
    form.setValue("idpEntityId", idpEntityId);
  }, [parseSAMLMetadataMutation, metadataUrl, form]);

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        <Button variant="outline">Edit</Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="w-full space-y-6"
          >
            <AlertDialogHeader>
              <AlertDialogTitle>
                Edit identity provider configuration
              </AlertDialogTitle>
            </AlertDialogHeader>

            <FormItem>
              <Label htmlFor={id}>IDP Metadata URL</Label>
              <div className="flex w-full items-center space-x-2">
                <Input
                  id={id}
                  placeholder="https://identity-provider.com/app/123/metadata.xml"
                  value={metadataUrl}
                  onChange={(e) => setMetadataUrl(e.target.value)}
                />
                <Button type="button" onClick={handleLoadMetadata}>
                  Load from metadata
                </Button>
              </div>
              <FormDescription>IDP Metadata URL.</FormDescription>
            </FormItem>

            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-background px-2 text-muted-foreground">
                  Or manually enter
                </span>
              </div>
            </div>

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
              render={({ field: { onChange } }) => (
                <FormItem>
                  <FormLabel>IDP Certificate</FormLabel>
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
                    IDP Certificate, as a PEM-encoded X.509 certificate. These
                    start with '-----BEGIN CERTIFICATE-----' and end with
                    '-----END CERTIFICATE-----'.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

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
