import { AnalyticsBrowser } from "@segment/analytics-next";
import { POSTHOG_API_KEY, SEGMENT_WRITE_KEY } from "@/config";
import posthog from "posthog-js";

export const analytics = AnalyticsBrowser.load({
  writeKey: SEGMENT_WRITE_KEY,
});

posthog.init(POSTHOG_API_KEY, {
  api_host: "https://us.i.posthog.com",
  segment: analytics as any,
  capture_pageview: false,
  person_profiles: "identified_only",
});
