import {
  HttpTransport,
  PublicClient,
  generateSmartAccount,
  waitTillCompleted,
} from "@nilfoundation/niljs";
import { TutorialChecksStatus, setTutorialChecksState } from "../../../pages/tutorials/model";
import { $rpcUrl } from "../../account-connector/model";
import { $contracts, deploySmartContractFx } from "../../contracts/models/base";
import { setCompletedTutorial } from "../../tutorial/model";
import { tutorialContractStepFailedEvent, tutorialContractStepPassedEvent } from "../model";

async function runTutorialCheckFour() {
  const client = new PublicClient({
    transport: new HttpTransport({
      endpoint: $rpcUrl.getState(),
    }),
    shardId: 1,
  });

  const counterContract = $contracts.getState().find((contract) => contract.name === "Counter")!;
  const deployerContract = $contracts.getState().find((contract) => contract.name === "Deployer")!;

  const appCounter = {
    name: "Counter",
    bytecode: counterContract.bytecode,
    abi: counterContract.abi,
    sourcecode: counterContract.sourcecode,
  };

  const appDeployer = {
    name: "Deployer",
    bytecode: deployerContract.bytecode,
    abi: deployerContract.abi,
    sourcecode: deployerContract.sourcecode,
  };

  const smartAccount = await generateSmartAccount({
    shardId: 1,
    rpcEndpoint: $rpcUrl.getState(),
    faucetEndpoint: $rpcUrl.getState(),
  });

  tutorialContractStepPassedEvent("A new smart account has been generated!");

  const resultDeployer = await deploySmartContractFx({
    app: appDeployer,
    args: [],
    shardId: 2,
    smartAccount,
  });

  tutorialContractStepPassedEvent("Deployer has been deployed!");

  const gasPrice = await client.getGasPrice(1);

  const hashDeploy = await smartAccount.sendTransaction({
    to: resultDeployer.address,
    abi: deployerContract.abi,
    functionName: "deploy",
    args: [appCounter.bytecode],
    feeCredit: gasPrice * 500_000n,
  });

  const resDeploy = await waitTillCompleted(client, hashDeploy);

  const checkDeploy = await resDeploy.some((receipt) => !receipt.success);

  if (checkDeploy) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    console.log(resDeploy);
    tutorialContractStepFailedEvent("Failed to call Deployer.deploy()!");
    return;
  }

  tutorialContractStepPassedEvent("Counter has been deployed!");

  const counterAddress = resDeploy.at(2)?.contractAddress as `0x${string}`;

  const hashIncrement = await smartAccount.sendTransaction({
    to: counterAddress,
    abi: counterContract.abi,
    functionName: "increment",
    args: [],
    feeCredit: gasPrice * 500_000n,
  });

  const resIncrement = await waitTillCompleted(client, hashIncrement);

  const checkIncrement = resIncrement.some((receipt) => !receipt.success);

  if (checkIncrement) {
    setTutorialChecksState(TutorialChecksStatus.Failed);
    console.log(resIncrement);
    tutorialContractStepFailedEvent("Failed to call Counter.increment()!");
    return;
  }

  tutorialContractStepPassedEvent("Counter.increment() has been called successfully!");

  setTutorialChecksState(TutorialChecksStatus.Successful);

  tutorialContractStepPassedEvent("Tutorial has been completed successfully!");

  setCompletedTutorial(3);
}

export default runTutorialCheckFour;
