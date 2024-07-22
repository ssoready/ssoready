import { AnalyticsBrowser } from "@segment/analytics-next";
import posthog from "posthog-js";
import { useConfig } from "@/config";
import React, {
  createContext,
  ReactNode,
  useContext,
  useEffect,
  useState,
} from "react";

const Context = createContext<AnalyticsBrowser>(undefined as any);

export function AnalyticsProvider({ children }: { children?: ReactNode }) {
  const { SEGMENT_WRITE_KEY, POSTHOG_API_KEY } = useConfig();
  const [analytics, setAnalytics] = useState<AnalyticsBrowser | undefined>();

  useEffect(() => {
    const analytics = AnalyticsBrowser.load({
      writeKey: SEGMENT_WRITE_KEY ?? "",
    });

    posthog.init(POSTHOG_API_KEY ?? "", {
      api_host: "https://us.i.posthog.com",
      segment: analytics as any,
      capture_pageview: false,
      person_profiles: "identified_only",
    });

    setAnalytics(analytics);
  }, [SEGMENT_WRITE_KEY, POSTHOG_API_KEY]);

  if (!analytics) {
    return;
  }

  return <Context.Provider value={analytics}>{children}</Context.Provider>;
}

export function useAnalytics(): AnalyticsBrowser {
  return useContext(Context);
}
