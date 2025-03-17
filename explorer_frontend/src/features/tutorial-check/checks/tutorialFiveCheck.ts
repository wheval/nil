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

async function runTutorialCheckFive() {
  const client = new PublicClient({
    transport: new HttpTransport({
      endpoint: $rpcUrl.getState(),
    }),
    shardId: 1,
  });

  const gasPrice = await client.getGasPrice(1);

  const receiverContract = $contracts.getState().find((contract) => contract.name === "Receiver")!;
  const NFTContract = $contracts.getState().find((contract) => contract.name === "NFT")!;

  const appReceiver: App = {
    name: "Receiver",
    bytecode: receiverContract.bytecode,
    abi: receiverContract.abi,
    sourcecode: receiverContract.sourcecode,
  };

  const appNFT: App = {
    name: "NFT",
    bytecode: NFTContract.bytecode,
    abi: NFTContract.abi,
    sourcecode: NFTContract.sourcecode,
  };

  const smartAccount = await generateSmartAccount({
    shardId: 1,
    rpcEndpoint: $rpcUrl.getState(),
    faucetEndpoint: $rpcUrl.getState(),
  });

  tutorialContractStepPassedEvent("A new smart account has been generated!");

  const resultReceiver = await deploySmartContractFx({
    app: appReceiver,
    args: [],
    shardId: 1,
    smartAccount,
  });

  const resultNFT = await deploySmartContractFx({
    app: appNFT,
    args: [],
    shardId: 2,
    smartAccount,
  });

  tutorialContractStepPassedEvent("Receiver and NFT have been deployed successfully!");

  const mintRequest = await smartAccount.sendTransaction({
    to: resultNFT.address,
    abi: NFTContract.abi,
    functionName: "mintNFT",
    args: [],
    feeCredit: gasPrice * 500_000n,
  });

  const resMinting = await waitTillCompleted(client, mintRequest);

  const checkMinting = await resMinting.some((receipt) => !receipt.success);

  if (checkMinting) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    console.log(resMinting);
    tutorialContractStepFailedEvent("Failed to mint the NFT!");
    return;
  }

  tutorialContractStepPassedEvent("NFT has been minted successfully!");

  const secondMintRequest = await smartAccount.sendTransaction({
    to: resultNFT.address,
    abi: NFTContract.abi,
    functionName: "mintNFT",
    args: [],
  });

  const resSecondMinting = await waitTillCompleted(client, secondMintRequest);

  const checkSecondMinting = await resSecondMinting.some((receipt) => !receipt.success);

  if (!checkSecondMinting) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    console.log(resSecondMinting);
    tutorialContractStepFailedEvent("NFT has been minted twice!");
    return;
  }

  tutorialContractStepPassedEvent("NFT is protected against repeated minting!");

  const sendRequest = await smartAccount.sendTransaction({
    to: resultNFT.address,
    abi: NFTContract.abi,
    functionName: "sendNFT",
    args: [resultReceiver.address],
    feeCredit: gasPrice * 500_000n,
  });

  const resSending = await waitTillCompleted(client, sendRequest);

  const checkSending = await resSending.some((receipt) => !receipt.success);

  if (checkSending) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    console.log(resSending);
    tutorialContractStepFailedEvent("Failed to send the NFT!");
    return;
  }

  const result = await client.getTokens(resultReceiver.address, "latest");

  if (Object.keys(result).length === 0) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    tutorialContractStepFailedEvent("NFT has not been received!");
    return;
  }

  tutorialContractStepPassedEvent("NFT has been received successfully!");

  setTutorialChecksState(TutorialChecksStatus.Successful);

  tutorialContractStepPassedEvent("Tutorial has been completed successfully!");

  setCompletedTutorial(5);
}

export default runTutorialCheckFive;
