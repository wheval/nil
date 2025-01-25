
import { LangChainAdapter } from 'ai';
import handler from '../../lib/handler';
import { NextResponse } from 'next/server';
import axios from "axios";

export const maxDuration = 60;

export async function POST(req: Request) {
  if (req.method !== "POST") {
    return new Response(
      JSON.stringify({ message: "Only POST requests allowed" }),
      { status: 405 },
    );
  }
  let result;
  const data = await req.json();
  const { messages, token } = data;

  const secretKey = process.env.RECAPTCHA_SECRET_KEY;

  if (!token) {
    return new Response(JSON.stringify({ message: "Token not found" }), {
      status: 405,
    });
  }

  try {
    const response = await axios.post(
      `https://www.google.com/recaptcha/api/siteverify?secret=${secretKey}&response=${token}`,
    );

    if (response.data.success) {
      const query = messages.at(-1).content;

      const handlerToUse = await handler();
      const retrievedDocs = await handlerToUse.vectorRetriever.invoke(query);

      const sources = retrievedDocs.map((doc) => {
        if (doc.metadata.source.endsWith(".sol")) {
          return "https://www.npmjs.com/package/@nilfoundation/smart-contracts";
        } else {
          return doc.metadata.source;
        }
      });
      const intent = await handlerToUse.intentLllmChain.invoke({
        query: query
      });
      if (intent == "Question") {
        result = await handlerToUse.genericLllmChain.stream({
          query: query,
          context: retrievedDocs,
          sources: sources
        });
      } else {
        result = await handlerToUse.generatorLllmChain.stream({
          query: query,
          context: retrievedDocs,
          sources: sources
    })
    }
      const resFinal = LangChainAdapter.toDataStreamResponse(result);
      resFinal.headers.append("Access-Control-Allow-Origin", "*");
      resFinal.headers.append("Access-Control-Allow-Methods", "GET,OPTIONS,PATCH,DELETE,POST,PUT");
      return resFinal;

    } else {
      return new Response(JSON.stringify({ message: "Failed to verify" }), {
        status: 405,
        headers: {
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET,OPTIONS,PATCH,DELETE,POST,PUT"
        }
      });
    }
  } catch (error) {
    console.log(error);
    return new Response(JSON.stringify({ message: "Internal server error" }), {
      status: 500,
      headers: {
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET,OPTIONS,PATCH,DELETE,POST,PUT"
        }
    });
}}

export async function OPTIONS(req: Request) {
  return new NextResponse(null, {
    status: 200,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET,OPTIONS,PATCH,DELETE,POST,PUT',
      'Access-Control-Allow-Headers': 'X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version',
    }
  })
}
