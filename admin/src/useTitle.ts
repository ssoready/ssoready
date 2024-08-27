import { useQuery } from "@connectrpc/connect-query";
import { adminWhoami } from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";

export function useTitle(title: string): string {
  const { data: whoami } = useQuery(adminWhoami, {});

  if (whoami?.adminApplicationName) {
    return `${title} | ${whoami.adminApplicationName} Settings Panel`;
  }
  return `${title} | Settings Panel`;
}
