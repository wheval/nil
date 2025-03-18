import {
  HttpTransport,
  PublicClient,
  generateSmartAccount,
  waitTillCompleted,
} from "@nilfoundation/niljs";
import { TutorialChecksStatus, setTutorialChecksState } from "../../../pages/tutorials/model";
import { $rpcUrl } from "../../account-connector/model";
import type { App } from "../../code/types";
import { $contracts, deploySmartContractFx } from "../../contracts/models/base";
import { setCompletedTutorial } from "../../tutorial/model";
import { tutorialContractStepFailedEvent, tutorialContractStepPassedEvent } from "../model";

async function runTutorialCheckThree() {
  const client = new PublicClient({
    transport: new HttpTransport({
      endpoint: $rpcUrl.getState(),
    }),
    shardId: 1,
  });

  const requesterContract = $contracts
    .getState()
    .find((contract) => contract.name === "Requester")!;
  const requestedContract = $contracts
    .getState()
    .find((contract) => contract.name === "RequestedContract")!;

  const appRequester: App = {
    name: "Requester",
    bytecode: requesterContract.bytecode,
    abi: requesterContract.abi,
    sourcecode: requesterContract.sourcecode,
  };

  const appRequestedContract: App = {
    name: "RequestedContract",
    bytecode: requestedContract.bytecode,
    abi: requestedContract.abi,
    sourcecode: requestedContract.sourcecode,
  };

  const smartAccount = await generateSmartAccount({
    shardId: 1,
    rpcEndpoint: $rpcUrl.getState(),
    faucetEndpoint: $rpcUrl.getState(),
  });

  tutorialContractStepPassedEvent("A new smart account has been generated!");

  const resultRequester = await deploySmartContractFx({
    app: appRequester,
    args: [],
    shardId: 1,
    smartAccount,
  });

  const resultRequestedContract = await deploySmartContractFx({
    app: appRequestedContract,
    args: [],
    shardId: 2,
    smartAccount,
  });

  tutorialContractStepPassedEvent("Requester and RequestedContract have been deployed!");

  const hashRequest = await smartAccount.sendTransaction({
    to: resultRequester.address,
    abi: requesterContract.abi,
    functionName: "requestMultiplication",
    args: [resultRequestedContract.address],
  });

  const resRequest = await waitTillCompleted(client, hashRequest);

  const checkRequest = await resRequest.some((receipt) => !receipt.success);

  if (checkRequest) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    console.log(resRequest);
    tutorialContractStepFailedEvent("Failed to call Requester.requestMultiplication()!");
    return;
  }

  tutorialContractStepPassedEvent(
    "Requester.requestMultiplication() has been called successfully!",
  );

  const result = await client.call(
    {
      to: resultRequester.address,
      abi: requesterContract.abi,
      functionName: "getResult",
    },
    "latest",
  );

  if (!result.decodedData) {
    tutorialContractStepFailedEvent("Failed to verify the result of multiplication!");
    setTutorialChecksState(TutorialChecksStatus.Failed);
    return;
  }

  tutorialContractStepPassedEvent("The result of multiplication has been verified!");

  setTutorialChecksState(TutorialChecksStatus.Successful);

  setCompletedTutorial(3);

  tutorialContractStepPassedEvent("Tutorial Three has been completed successfully!");
}

export default runTutorialCheckThree;
