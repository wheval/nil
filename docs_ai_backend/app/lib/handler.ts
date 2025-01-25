import { DBClient } from "../../src/clients/DBClient.ts";
import { QueryHandler } from "../../src/core/QueryHandler.ts";

const handler = async () => {
  const dbClient = new DBClient();
  const retriever = await dbClient.provideVectorStore();
  const handler = new QueryHandler(retriever);
  await handler.createLLMsAndChains();
  return handler;
}

export default handler;