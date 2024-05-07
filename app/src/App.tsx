import React from "react";
import { Outlet, Route, Routes } from "react-router";
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

const queryClient = new QueryClient();

const transport = createConnectTransport({
  baseUrl: "http://localhost:8081/internal/connect",
  interceptors: [
    (next) => async (req) => {
      req.header.set("Authorization", `Bearer ${getSessionToken()}`);
      return next(req);
    },
  ],
});

export function App() {
  return (
    <TransportProvider transport={transport}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/verify-email" element={<VerifyEmailPage />} />

            <Route path="/" element={<LoginGate />}>
              <Route path="/" element={<Page />}>
                <Route path="/" element={<HomePage />} />
                <Route
                  path="/environments/new"
                  element={<CreateEnvironmentPage />}
                />
                <Route
                  path="/environments/:environmentId"
                  element={<ViewEnvironmentPage />}
                />
                <Route
                  path="/environments/:environmentId/api-keys/:apiKeyId"
                  element={<ViewAPIKeyPage />}
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
              </Route>
            </Route>
          </Routes>
        </BrowserRouter>
      </QueryClientProvider>
    </TransportProvider>
  );
}
