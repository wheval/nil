import "@nomicfoundation/hardhat-toolbox";
import "@nilfoundation/hardhat-nil-plugin";
import * as dotenv from "dotenv";
import type { HardhatConfig } from "hardhat/types";

// tasks
import "./tasks/deploy-incrementer";

dotenv.config();

const config: HardhatConfig = {
  ignition: {
    requiredConfirmations: 1,
  },
  defaultNetwork: "nil",
  solidity: {
    version: "0.8.26", // or your desired version
    settings: {
      viaIR: true, // needed to compile router
      optimizer: {
        enabled: true,
        runs: 200,
      },
    },
  },
  networks: {
    nil: {
      nil: true, // needed to externally mark nil network
      url: process.env.NIL_RPC_ENDPOINT,
      accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : [],
    },
  },
};

export default config;
