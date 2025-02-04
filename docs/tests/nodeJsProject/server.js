//startImport
const { createServer } = require("node:http");
const { generateSmartAccount } = require("@nilfoundation/niljs");

const hostname = "127.0.0.1";
const port = 3000;
//endImport

const RPC_ENDPOINT = "http://127.0.0.1:8529";
const FAUCET_ENDPOINT = "http://127.0.0.1:8529";

//startServer
const server = createServer((req, res) => {
  (async () => {
    try {
      res.statusCode = 200;
      res.setHeader("Content-Type", "text/plain");

      const smartAccount = await generateSmartAccount({
        shardId: 1,
        rpcEndpoint: RPC_ENDPOINT,
        faucetEndpoint: FAUCET_ENDPOINT,
      });

      const smartAccountAddress = smartAccount.address;

      res.write(`New smart account address: ${smartAccountAddress}\n`);
      res.on("finish", () => {
        console.log(`New smart account address: ${smartAccountAddress}`);
      });
      res.end();
    } catch (error) {
      console.error(error);
      res.statusCode = 500;
      res.write("An error occurred while creating the smart account.");
      res.end();
    }
  })();
});

server.listen(port, hostname, () => {
  console.log(`Server running at http://${hostname}:${port}/`);
});
//endServer
