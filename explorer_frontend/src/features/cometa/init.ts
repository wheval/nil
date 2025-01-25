import { CometaService, HttpTransport } from "@nilfoundation/niljs";
import { combine, sample } from "effector";
import { $endpoint } from "../account-connector/model";
import {
  $cometaService,
  $commetaEndpoint,
  createCometaService,
  createCometaServiceFx,
} from "./model";

const $refinedEndpoint = combine(
  $endpoint,
  $commetaEndpoint,
  (endpoint, customEndpoint) => endpoint || customEndpoint,
);

$refinedEndpoint.watch((endpoint) => {
  if (endpoint) {
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
  source: $refinedEndpoint,
  target: createCometaServiceFx,
});

createCometaService();
