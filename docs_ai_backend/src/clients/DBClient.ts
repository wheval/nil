import { createClient } from "@libsql/client";
import { LibSQLVectorStore } from "@langchain/community/vectorstores/libsql";
import { OpenAIEmbeddings } from "@langchain/openai";
import getConfig from 'next/config';

export class DBClient {
  openAIEmbeddings: OpenAIEmbeddings;

  constructor() {
    this.openAIEmbeddings = new OpenAIEmbeddings({
      model: "text-embedding-3-large",
      apiKey: process.env.OPENAI_API_KEY || "default",
    });
  }

  public async provideVectorStore() {
    const db = createClient({
      url: `http://${process.env.DB_URL}`,
    });


    const vectorStore = new LibSQLVectorStore(this.openAIEmbeddings, {
      db: db,
      table: "EMBEDDINGS_DOCS",
      column: "EMBEDDING_COLUMN",
    });

    return vectorStore.asRetriever();
  }


}
