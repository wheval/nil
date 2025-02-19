import { CometaService, HttpTransport } from "@nilfoundation/niljs";
import { combine, sample } from "effector";
import { $rpcUrl } from "../account-connector/model";
import { $cometaApiUrl, $cometaService, createCometaService, createCometaServiceFx } from "./model";

const $refinedCometaApiUrl = combine(
  $rpcUrl,
  $cometaApiUrl,
  (rpcUrl, customCometaApiUrl) => rpcUrl || customCometaApiUrl,
);

$refinedCometaApiUrl.watch((url) => {
  if (url) {
    createCometaService();
  }
});

createCometaServiceFx.use(async (endpoint) => {
  const cometaService = new CometaService({
    transport: new HttpTransport({ endpoint }),
  });

  return cometaService;
});

$cometaService.on(createCometaServiceFx.doneData, (_, cometaService) => cometaService);

sample({
  clock: createCometaService,
  source: $refinedCometaApiUrl,
  target: createCometaServiceFx,
});

createCometaService();
