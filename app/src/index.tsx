import React from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import hljs from "highlight.js/lib/core";
import bash from "highlight.js/lib/languages/bash";
import xml from "highlight.js/lib/languages/xml";
import json from "highlight.js/lib/languages/json";

import * as Sentry from "@sentry/react";
import { SENTRY_DSN, SENTRY_ENVIRONMENT } from "@/config";

Sentry.init({
  dsn: SENTRY_DSN,
  environment: SENTRY_ENVIRONMENT,
  integrations: [
    Sentry.browserTracingIntegration(),
    Sentry.replayIntegration(),
    Sentry.httpClientIntegration(),
  ],
  replaysSessionSampleRate: 1.0,
  replaysOnErrorSampleRate: 1.0,
  debug: true,
});

hljs.registerLanguage("bash", bash);
hljs.registerLanguage("xml", xml);
hljs.registerLanguage("json", json);

const root = createRoot(document.getElementById("react-root")!);
root.render(<App />);
