import { FAUCET_GLOBAL, RPC_GLOBAL } from "./globals";

//startImportStatements
import { HttpTransport, PublicClient, generateSmartAccount, topUp } from "@nilfoundation/niljs";
//endImportStatements

const RPC_ENDPOINT = RPC_GLOBAL;
const FAUCET_ENDPOINT = FAUCET_GLOBAL;

describe.sequential("Nil.js can use the faucet service", async () => {
  test.sequential("Nil.js can use the faucet service to do a default token top-up", async () => {
    //startDefaultExample
    const client = new PublicClient({
      transport: new HttpTransport({
        endpoint: RPC_ENDPOINT,
      }),
      shardId: 1,
    });

    const smartAccount = await generateSmartAccount({
      shardId: 1,
      rpcEndpoint: RPC_ENDPOINT,
      faucetEndpoint: FAUCET_ENDPOINT,
    });

    const resultBeforeTopUp = await client.getBalance(smartAccount.address);

    console.log(resultBeforeTopUp);

    //endDefaultExample

    //startContDefaultExample

    await topUp({
      address: smartAccount.address,
      faucetEndpoint: FAUCET_ENDPOINT,
      rpcEndpoint: RPC_ENDPOINT,
      token: "NIL",
      amount: 1_000_000,
    });

    const result = await client.getBalance(smartAccount.address);

    console.log(result);

    //endContDefaultExample

    expect(result > resultBeforeTopUp);
  });

  test.sequential(
    "Nil.js can use the faucet service to handle custom tokens top-up",
    async () => {
      //startBTCExample
      const client = new PublicClient({
        transport: new HttpTransport({
          endpoint: RPC_ENDPOINT,
        }),
        shardId: 1,
      });

      const smartAccount = await generateSmartAccount({
        shardId: 1,
        rpcEndpoint: RPC_ENDPOINT,
        faucetEndpoint: FAUCET_ENDPOINT,
      });

      const resultBeforeTopUp = await client.getTokens(smartAccount.address, "latest");

      console.log(resultBeforeTopUp);

      //endBTCExample

      //startContBTCExample

      await topUp({
        address: smartAccount.address,
        faucetEndpoint: FAUCET_ENDPOINT,
        rpcEndpoint: RPC_ENDPOINT,
        token: "BTC",
        amount: 1_000_000,
      });

      const result = await client.getTokens(smartAccount.address, "latest");

      console.log(result);

      //endContBTCExample

      expect(Object.values(result)).toContain(1_000_000n);
    },
    40000,
  );
});
