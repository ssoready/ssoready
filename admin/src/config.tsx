import React, {
  createContext,
  ReactNode,
  useContext,
  useEffect,
  useState,
} from "react";

export type Config = {
  SENTRY_DSN?: string;
  SENTRY_ENVIRONMENT?: string;
  SEGMENT_WRITE_KEY?: string;
  POSTHOG_API_KEY?: string;
  API_URL: string;
};

const Context = createContext<Config>(undefined as any);

export function ConfigProvider({ children }: { children?: ReactNode }) {
  const [config, setConfig] = useState<Config | undefined>();

  useEffect(() => {
    (async () => {
      const res = await (await fetch("/config.json")).json();
      setConfig(res);
    })();
  }, []);

  if (!config) {
    return;
  }

  return <Context.Provider value={config}>{children}</Context.Provider>;
}

export function useConfig(): Config {
  return useContext(Context);
}
