import React, { useCallback, useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import { getEnvironmentCustomDomainSettings } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import { Badge } from "@/components/ui/badge";
import { ArrowRightIcon, CopyIcon } from "lucide-react";
import { offset, useFloating, useTransitionStyles } from "@floating-ui/react";

export function ViewCustomDomainsPage() {
  const { environmentId } = useParams();
  const { data: customDomainsSettings } = useQuery(
    getEnvironmentCustomDomainSettings,
    {
      environmentId,
    },
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          Custom Domain for{" "}
          <code className="text-xl bg-gray-100 py-1 px-2 rounded-sm">
            auth.ssoready.com
          </code>
        </CardTitle>
        <CardDescription>
          SAML and SCIM both require that you set up a server that your
          customer's Identity Provider talks to. SSOReady runs that server for
          you. With a custom domain, your customers don't see any URLs with
          "ssoready.com" in the name.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-[1fr_auto_1fr] gap-x-4">
          <div className="text-sm font-semibold mb-2">DNS record name</div>
          <div></div>
          <div className="text-sm font-semibold mb-2">DNS record value</div>
          <div className="flex items-center text-sm border rounded">
            <div className="py-2 px-4 bg-muted">CNAME</div>
            <div className="flex-1 py-2 px-4">
              {customDomainsSettings?.customAuthDomain}
            </div>
            <div className="px-2">
              {customDomainsSettings && (
                <CopyButton value={customDomainsSettings.customAuthDomain} />
              )}
            </div>
          </div>
          <div className="flex items-center">
            <ArrowRightIcon className="h-5 w-5 text-muted-foreground" />
          </div>
          <div className="flex items-center text-sm border rounded">
            <div className="flex-1 py-2 px-4">
              {customDomainsSettings?.customAuthDomainCnameValue}
            </div>
            <div className="px-2">
              {customDomainsSettings && (
                <CopyButton
                  value={customDomainsSettings.customAuthDomainCnameValue}
                />
              )}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function CopyButton({ value }: { value: string }) {
  const [open, setOpen] = useState(false);
  const { refs, floatingStyles, context } = useFloating({
    open,
    onOpenChange: setOpen,
    placement: "top",
    middleware: [offset(5)],
  });
  const { isMounted, styles } = useTransitionStyles(context, {
    duration: 150,
    initial: { opacity: 0, transform: "translateY(0)" },
    open: { opacity: 1, transform: "translateY(-5px)" },
  });

  useEffect(() => {
    if (open) {
      const timeoutId = setTimeout(() => {
        setOpen(false);
      }, 1000);
      return () => clearTimeout(timeoutId);
    }
  }, [open]);

  const handleCopy = useCallback(async () => {
    await navigator.clipboard.writeText(value);
    setOpen(true);
  }, [value, setOpen]);

  return (
    <>
      <CopyIcon
        ref={refs.setReference}
        onClick={handleCopy}
        className="cursor-pointer text-muted-foreground hover:text-foreground h-4 w-4"
      />

      {open && (
        <div ref={refs.setFloating} style={floatingStyles}>
          {isMounted && (
            <div
              style={styles}
              className="font-sans bg-black text-white px-2 py-1 text-xs rounded"
            >
              Copied!
            </div>
          )}
        </div>
      )}
    </>
  );
}
