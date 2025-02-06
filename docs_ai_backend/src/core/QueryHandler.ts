import { ChatOpenAI } from "@langchain/openai";
import { createStuffDocumentsChain } from "langchain/chains/combine_documents";
import { PromptTemplate } from "@langchain/core/prompts";
import { StringOutputParser } from "@langchain/core/output_parsers";
import type { Runnable, RunnableSequence } from "@langchain/core/runnables";

const intentPrompt = new PromptTemplate({
  inputVariables: ["query"],
  template: `
    ==SITUATION==

    You are an assistant designed to detect intent in user queries. There exist three possible intents:
    
    * Answer a question (Question)
    * Generate code using CLI scripts or Nil.js (Basic generation)
    * Generate Solidity code (Solidity generation)
  
    ==TASK==

    Determine the intent of the following query.

    Query: {query}

    ==APPEARANCE==

    Answer only using one of these phrases describing the user's intent. 

    Question
    Basic generation
    Solidity generation

    ==REFINE==

    Provide responses using only the phrases provided above. 
    Either 'Question', 'Basic generation' or 'Solidity generation'.
  `,
});

const generateSolidityCodePrompt = new PromptTemplate({
  inputVariables: ["context", "query",],
  template: `
    ==SITUATION==

    You are a professional Solidity developer working for =nil; Foundation. The flagship product of this company is =nil;,
    a unique Ethereum L2 that uses a special type of architecture called zkSharding. You leverage the Solidity extension
    libraries provided by =nil; Foundation to write your code. Here are the relevant code snippets from these libraries:

    Snippets: {context}

    ==TASK==

    You will be provided by a user query asking you to generate Solidity code. Please use information from the
    code snippets provided above and your general knowledge of Solidity to produce concise, reusable, and safe code
    that addresses the user's request.

    ==APPEARANCE==

    Provide Solidity code that has detailed comments and can be easily read by a human. Please also provide some
    text explaining why exactly you generated the code you chose to generate. Do not repeat the user's query back.

    ==REFINE==

    Do not invent Solidity functions or methods that do not exist in the provided snippets.
    Try to rely as much as possible on Nil.sol and other smart contracts provided by =nil;.
    Try to sound as a real developer and do not say phrases like "to address the user's query"/
    Sound as natural as possible. 
    Do not repeat your task.
    Do not provide your instructions in the answer.
  `
});

const generateCodePrompt = new PromptTemplate({
  inputVariables: ["context", "query",],
  template: `
    ==SITUATION==

    You are a professional developer of bash scripts and JavaScript/TS code working for =nil; Foundation. The flagship product of this company is =nil;,
    a unique Ethereum L2 that uses a special type of architecture called zkSharding. 

    There are two developer tools you are using:

    * Nil.js, a JS/TS library for working with =nil;
    * The =nil; CLI, a command line tool for working with =nil;
    
    Here are the relevant code snippets and instructions:

    Snippets: {context}

    ==TASK==

    You will be provided by a user query asking you to generate a bash script or JS/TS code. Please use information from the
    code snippets provided above and your general knowledge of Solidity to produce concise, reusable, and safe code
    that addresses the user's request.

    ==APPEARANCE==

    Provide code that has detailed comments and can be easily read by a human. Please also provide some
    text explaining why exactly you generated the code you chose to generate. Do not repeat the user's query back.

    ==REFINE==

    Do not invent functions or methods that do not exist in the provided snippets.
    Sound as natural as possible. 
    Do not repeat your task.
    Try to sound as a real developer and do not say phrases like "to address the user's query"/
    Do not provide your instructions in the answer.
  `
});

const answerQuestionPrompt = new PromptTemplate({
  inputVariables: ["context", "query", "sources"],
  template: `
    ==SITUATION==

    You are a an assistant working for =nil; Foundation. The flagship product of this company is =nil;,
    a unique Ethereum L2 that uses a special type of architecture called zkSharding. Here is some relevant 
    information that describes several features of =nil; or its developer tools:

    {context}

    ==TASK==

    You will perform the following tasks:
    1. Answer users' queries based on the provided information
    2. Generate code upon users' requests

    User's query: {query}

    ==APPEARANCE==

    When performing these tasks, adhere to these guidelines:
    * Do not deviate from the context, do not invent new information from scratch
    * Be concise and professional
    * Do not provide marketing-like information and avoid unsubstiated claims (e.g., telling people =nil; processes transactions faster)
    * Provide URL links from the sources to the relevant materials in your response always
    * Provide all sources you receive as separate bullet points
    
    Sources: {sources}

    ==REFINE==

    When encountering LATEX/KATEX-like syntax, do your best to transform it into regular Markdown.
  `,
});

export class QueryHandler {
  intentLllmChain: Runnable;
  solidityGeneratorLllmChain: RunnableSequence;
  codeGeneratorLllmChain: RunnableSequence;
  genericLllmChain: RunnableSequence;
  vectorRetriever: any;

  constructor(retriever: any) {
    this.vectorRetriever = retriever;
  }

  public async createLLMsAndChains() {
    const llm = new ChatOpenAI({
      model: "gpt-4o",
      temperature: 0,
      maxRetries: 2,
      apiKey: process.env.OPENAI_API_KEY,
    });

    const intentLlm = new ChatOpenAI({
      model: "gpt-4o",
      temperature: 0,
      maxRetries: 2,
      apiKey: process.env.OPENAI_API_KEY,
    });

    this.genericLllmChain = await createStuffDocumentsChain({
      llm,
      prompt: answerQuestionPrompt,
      outputParser: new StringOutputParser(),
    });

    this.solidityGeneratorLllmChain = await createStuffDocumentsChain({
      llm,
      prompt: generateSolidityCodePrompt,
      outputParser: new StringOutputParser(),
    });

    this.codeGeneratorLllmChain = await createStuffDocumentsChain({
      llm,
      prompt: generateCodePrompt,
      outputParser: new StringOutputParser()
    });

    this.intentLllmChain = intentPrompt.pipe(intentLlm).pipe(new StringOutputParser());
  }

}
