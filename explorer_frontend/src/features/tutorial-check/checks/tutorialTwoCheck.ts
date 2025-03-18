import {
  HttpTransport,
  PublicClient,
  generateSmartAccount,
  waitTillCompleted,
} from "@nilfoundation/niljs";
import { TutorialChecksStatus, setTutorialChecksState } from "../../../pages/tutorials/model";
import type { App } from "../../../types";
import { $rpcUrl } from "../../account-connector/model";
import { $contracts, deploySmartContractFx } from "../../contracts/models/base";
import { setCompletedTutorial } from "../../tutorial/model";
import { tutorialContractStepFailedEvent, tutorialContractStepPassedEvent } from "../model";

const CUSTOM_TOKEN_AMOUNT = 30_000n;

async function runTutorialCheckTwo() {
  const client = new PublicClient({
    transport: new HttpTransport({
      endpoint: $rpcUrl.getState(),
    }),
    shardId: 1,
  });

  const operatorContract = $contracts.getState().find((contract) => contract.name === "Operator")!;

  const customTokenContract = $contracts
    .getState()
    .find((contract) => contract.name === "CustomToken")!;

  const appOperator: App = {
    name: "Operator",
    bytecode: operatorContract.bytecode,
    abi: operatorContract.abi,
    sourcecode: operatorContract.sourcecode,
  };

  const appCustomToken: App = {
    name: "CustomToken",
    bytecode: customTokenContract.bytecode,
    abi: customTokenContract.abi,
    sourcecode: customTokenContract.sourcecode,
  };

  const smartAccount = await generateSmartAccount({
    shardId: 1,
    rpcEndpoint: $rpcUrl.getState(),
    faucetEndpoint: $rpcUrl.getState(),
  });

  tutorialContractStepPassedEvent("A new smart account has been generated!");

  const resultOperator = await deploySmartContractFx({
    app: appOperator,
    args: [],
    shardId: 1,
    smartAccount,
  });

  const resultCustomToken = await deploySmartContractFx({
    app: appCustomToken,
    args: [resultOperator.address],
    shardId: 2,
    smartAccount,
  });

  tutorialContractStepPassedEvent("Operator and CustomToken have been deployed!");

  const hashMinting = await smartAccount.sendTransaction({
    to: resultOperator.address,
    abi: operatorContract.abi,
    functionName: "checkMintToken",
    args: [resultCustomToken.address, CUSTOM_TOKEN_AMOUNT],
  });

  const resMinting = await waitTillCompleted(client, hashMinting);

  const checkMinting = resMinting.some((receipt) => !receipt.success);

  if (checkMinting) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    console.log(resMinting);
    tutorialContractStepFailedEvent("Failed to call mintTokenCustom()!");
    return;
  }

  tutorialContractStepPassedEvent("mintTokenCustom() has been called successfully!");

  const customTokenBalance = await client.getTokens(resultCustomToken.address, "latest");

  if (Object.values(customTokenBalance).at(0) === CUSTOM_TOKEN_AMOUNT) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    tutorialContractStepFailedEvent("CustomToken failed to mint tokens!");
    return;
  }

  tutorialContractStepPassedEvent("CustomToken has minted tokens successfully!");

  const hashSending = await smartAccount.sendTransaction({
    to: resultOperator.address,
    abi: operatorContract.abi,
    functionName: "checkSendToken",
    args: [resultCustomToken.address, CUSTOM_TOKEN_AMOUNT],
  });

  const resSending = await waitTillCompleted(client, hashSending);

  const checkSending = resSending.some((receipt) => !receipt.success);

  if (checkSending) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    console.log(resMinting);
    tutorialContractStepFailedEvent("Failed to call sendTokenCustom()!");
    return;
  }

  tutorialContractStepPassedEvent("sendTokenCustom() has been called successfully!");

  const customTokenBalanceOperator = await client.getTokens(resultOperator.address, "latest");

  if (Object.values(customTokenBalanceOperator).at(1) === CUSTOM_TOKEN_AMOUNT) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    tutorialContractStepFailedEvent("Operator did not receive tokens from CustomToken!");
    return;
  }

  tutorialContractStepPassedEvent("Operator has received tokens from CustomToken successfully!");
  setTutorialChecksState(TutorialChecksStatus.Successful);
  tutorialContractStepPassedEvent("Tutorial has been completed successfully!");

  setCompletedTutorial(2);
}

export default runTutorialCheckTwo;
