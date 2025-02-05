import type {
  FaucetClient,
  LocalECDSAKeySigner,
  PublicClient,
  SmartAccountV1,
} from "@nilfoundation/niljs";
import {
  createClient,
  createFaucetClient,
  createSigner,
  initializeOrDeploySmartAccount,
} from "../../features/blockchain";
import { initializeActivities } from "../../features/store/model/activities.ts";
import {
  setFaucetClient,
  setPublicClient,
  setSigner,
} from "../../features/store/model/blockchain.ts";
import { setEndpoint } from "../../features/store/model/endpoint.ts";
import { setPrivateKey } from "../../features/store/model/privateKey.ts";
import {
  setExistingSmartAccount,
  setIsSmartAccountInitialized,
} from "../../features/store/model/smartAccount.ts";

// Saves other blockchain fields to Chrome storage
export async function saveUserDetails(fields: {
  rpcEndpoint: string;
  shardId: number;
  privateKey: string;
  smartAccountAddress: string;
}): Promise<void> {
  try {
    await chrome.storage.local.set({ blockchainFields: fields });
    console.log("Blockchain fields saved:", fields);
  } catch (error) {
    console.error("Error saving blockchain fields:", error);

    // Rethrow the error to propagate it to the caller
    throw new Error("Failed to save blockchain fields to Chrome storage");
  }
}

// Initializes blockchain resources by loading fields from Chrome storage and setting up clients
export async function initializeFromStorageAndSetup(): Promise<void> {
  try {
    const { blockchainFields } = await chrome.storage.local.get("blockchainFields");
    if (blockchainFields) {
      console.log("Loaded blockchain fields from storage: ", blockchainFields);

      const { rpcEndpoint, shardId, privateKey, smartAccountAddress } = blockchainFields;

      let publicClient: PublicClient;
      let signer: LocalECDSAKeySigner;
      let faucetClient: FaucetClient;
      let smartAccount: SmartAccountV1;

      try {
        publicClient = await createClient(rpcEndpoint, shardId);
      } catch (error) {
        console.error("Error creating PublicClient: ", error);
        throw new Error("Failed to create PublicClient");
      }

      try {
        signer = await createSigner(privateKey);
      } catch (error) {
        console.error("Error creating signer: ", error);
        throw new Error("Failed to create signer");
      }

      try {
        faucetClient = await createFaucetClient(rpcEndpoint);
      } catch (error) {
        console.error("Error creating FaucetClient: ", error);
        throw new Error("Failed to create FaucetClient");
      }

      try {
        smartAccount = await initializeOrDeploySmartAccount({
          client: publicClient,
          faucetClient: faucetClient,
          signer: signer,
          shardId: shardId,
          existingSmartAccountAddress: smartAccountAddress,
        });
      } catch (error) {
        console.error("Error initializing or deploying smartAccount: ", error);
        throw new Error("Failed to initialize or deploy smartAccount");
      }

      try {
        setIsSmartAccountInitialized(true);
        setExistingSmartAccount(smartAccount);
        setPrivateKey(privateKey);
        setEndpoint(rpcEndpoint);
        setPublicClient(publicClient);
        setSigner(signer);
        setFaucetClient(faucetClient);
        initializeActivities(smartAccount);
      } catch (error) {
        console.error("Error setting up application state: ", error);
        throw new Error("Failed to set up application state");
      }
    } else {
      console.log("No blockchain fields found in storage");
    }
  } catch (error) {
    console.error("Error initializing from storage and setting up blockchain resources: ", error);
    throw new Error(`Initialization failed: ${error instanceof Error ? error.message : error}`);
  }
}

export async function areBlockchainFieldsSet(): Promise<boolean> {
  try {
    const { blockchainFields } = await chrome.storage.local.get("blockchainFields");
    if (!blockchainFields) {
      return false;
    }

    const { rpcEndpoint, privateKey, shardId, smartAccountAddress } = blockchainFields;
    return Boolean(rpcEndpoint && privateKey && shardId && smartAccountAddress);
  } catch (error) {
    console.error("Error checking blockchain fields in storage: ", error);
    return false;
  }
}

// Clears blockchain-related fields from Chrome storage, retaining the private key
export async function clearState(): Promise<void> {
  try {
    // Retrieve existing blockchain fields
    const { blockchainFields } = await chrome.storage.local.get("blockchainFields");

    if (blockchainFields) {
      const { privateKey } = blockchainFields;

      // Retain only the private key in storage
      await chrome.storage.local.set({ blockchainFields: { privateKey } });
      console.log("Cleared blockchain fields except private key.");
    } else {
      console.log("No blockchain fields found in storage to clear.");
    }
  } catch (error) {
    console.error("Error clearing blockchain fields: ", error);
  }
}
