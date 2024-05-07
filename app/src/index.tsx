import React from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import hljs from "highlight.js/lib/core";
import xml from "highlight.js/lib/languages/xml";
import json from "highlight.js/lib/languages/json";

hljs.registerLanguage("xml", xml);
hljs.registerLanguage("json", json);

const root = createRoot(document.getElementById("react-root")!);
root.render(<App />);
