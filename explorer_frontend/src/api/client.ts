import type { AppRouter } from "@nilfoundation/explorer-backend";
import { createTRPCProxyClient, httpBatchLink } from "@trpc/client";
import { httpLink } from "@trpc/client/links/httpLink";
import { splitLink } from "@trpc/client/links/splitLink";
import { getRuntimeConfigOrThrow } from "../features/runtime-config";

const { API_REQUESTS_ENABLE_BATCHING, API_URL } = getRuntimeConfigOrThrow();
const url = API_URL || "/api";

export const client = createTRPCProxyClient<AppRouter>({
  links: [
    splitLink({
      condition() {
        return API_REQUESTS_ENABLE_BATCHING === "true";
      },
      true: httpBatchLink({ url }),
      false: httpLink({ url }),
    }),
  ],
});
