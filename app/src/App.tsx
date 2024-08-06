import React, { useEffect } from "react";
import { Outlet, Route, Routes, useLocation } from "react-router";
import { BrowserRouter } from "react-router-dom";
import { LoginPage } from "./pages/LoginPage";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TransportProvider } from "@connectrpc/connect-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { getSessionToken } from "./auth";
import { LoginGate } from "./components/LoginGate";
import { HomePage } from "./pages/HomePage";
import { Page } from "@/components/Page";
import { ViewEnvironmentPage } from "@/pages/ViewEnvironmentPage";
import { ViewOrganizationPage } from "@/pages/ViewOrganizationPage";
import { ViewSAMLConnectionPage } from "@/pages/ViewSAMLConnectionPage";
import { EditSAMLConnectionPage } from "@/pages/EditSAMLConnectionPage";
import { ViewSAMLFlowPage } from "@/pages/ViewSAMLFlowPage";
import { CreateEnvironmentPage } from "@/pages/CreateEnvironmentPage";
import { ViewAPIKeyPage } from "@/pages/ViewAPIKeyPage";
import { VerifyEmailPage } from "@/pages/VerifyEmailPage";
import { Toaster } from "@/components/ui/sonner";
import { OnboardingPage } from "@/pages/OnboardingPage";
import { NotFound } from "@/pages/NotFound";
import { ViewSAMLOAuthClientPage } from "@/pages/ViewSAMLOAuthClientPage";
import { ListAPIKeysPage } from "@/pages/ListAPIKeysPage";
import { AnalyticsProvider, useAnalytics } from "@/analytics";
import { Transport } from "@connectrpc/connect";
import { ConfigProvider, useConfig } from "@/config";
import * as Sentry from "@sentry/react";
import { MicrosoftCallbackPage } from "@/pages/internal/MicrosoftCallbackPage";
import { SettingsPage } from "@/pages/SettingsPage";
import { ViewSCIMDirectoryPage } from "@/pages/ViewSCIMDirectoryPage";

const queryClient = new QueryClient();

function useTransport(): Transport {
  const { API_URL } = useConfig();
  return createConnectTransport({
    baseUrl: API_URL,
    interceptors: [
      (next) => async (req) => {
        req.header.set("Authorization", `Bearer ${getSessionToken() ?? ""}`);
        return next(req);
      },
    ],
  });
}

export function AppRoutes() {
  const location = useLocation();
  const analytics = useAnalytics();
  useEffect(() => {
    analytics?.page();
  }, [analytics, location]);

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/verify-email" element={<VerifyEmailPage />} />
      <Route
        path="/internal/microsoft-callback"
        element={<MicrosoftCallbackPage />}
      />

      <Route path="/" element={<LoginGate />}>
        <Route path="/onboarding" element={<OnboardingPage />} />

        <Route path="/" element={<Page />}>
          <Route path="/" element={<HomePage />} />
          <Route path="/settings" element={<SettingsPage />} />

          <Route path="/environments/new" element={<CreateEnvironmentPage />} />
          <Route
            path="/environments/:environmentId"
            element={<ViewEnvironmentPage />}
          />
          <Route
            path="/environments/:environmentId/api-keys"
            element={<ListAPIKeysPage />}
          />
          <Route
            path="/environments/:environmentId/api-keys/:apiKeyId"
            element={<ViewAPIKeyPage />}
          />
          <Route
            path="/environments/:environmentId/saml-oauth-clients/:samlOAuthClientId"
            element={<ViewSAMLOAuthClientPage />}
          />
          <Route
            path="/environments/:environmentId/organizations/:organizationId"
            element={<ViewOrganizationPage />}
          />
          <Route
            path="/environments/:environmentId/organizations/:organizationId/saml-connections/:samlConnectionId"
            element={<ViewSAMLConnectionPage />}
          />
          <Route
            path="/environments/:environmentId/organizations/:organizationId/saml-connections/:samlConnectionId/flows"
            element={<ViewSAMLConnectionPage />}
          />
          <Route
            path="/environments/:environmentId/organizations/:organizationId/saml-connections/:samlConnectionId/edit"
            element={<EditSAMLConnectionPage />}
          />
          <Route
            path="/environments/:environmentId/organizations/:organizationId/saml-connections/:samlConnectionId/flows/:samlFlowId"
            element={<ViewSAMLFlowPage />}
          />
          <Route
            path="/environments/:environmentId/organizations/:organizationId/scim-directories/:scimDirectoryId"
            element={<ViewSCIMDirectoryPage />}
          />
          <Route
            path="/environments/:environmentId/organizations/:organizationId/scim-directories/:scimDirectoryId/groups"
            element={<ViewSCIMDirectoryPage />}
          />
        </Route>
      </Route>

      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}

export function AppWithinConfig() {
  const { SENTRY_DSN, SENTRY_ENVIRONMENT } = useConfig();
  const transport = useTransport();

  useEffect(() => {
    Sentry.init({
      dsn: SENTRY_DSN,
      environment: SENTRY_ENVIRONMENT,
      integrations: [
        Sentry.browserTracingIntegration(),
        Sentry.replayIntegration(),
        Sentry.httpClientIntegration(),
      ],
      replaysSessionSampleRate: 0.0,
      replaysOnErrorSampleRate: 1.0,
    });
  }, [SENTRY_DSN, SENTRY_ENVIRONMENT]);

  return (
    <>
      <TransportProvider transport={transport}>
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            <AppRoutes />
          </BrowserRouter>
        </QueryClientProvider>
      </TransportProvider>
      <Toaster />
    </>
  );
}

export function App() {
  return (
    <ConfigProvider>
      <AnalyticsProvider>
        <AppWithinConfig />
      </AnalyticsProvider>
    </ConfigProvider>
  );
}
