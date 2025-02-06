
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
      const pastMessages = messages.slice(-11,-1);

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

      const stream = new TransformStream();
      const writer = stream.writable.getWriter();

      (async () => {
      try {
        let chain;
        switch (intent) {
          case "Question":
            chain = handlerToUse.genericLllmChain;
            break;
          case "Basic generation":
            chain = handlerToUse.codeGeneratorLllmChain;
            break;
          case "Solidity generation":
            chain = handlerToUse.solidityGeneratorLllmChain;
            break;
        }
        const result = await chain.stream({
          query: query,
          context: retrievedDocs,
          sources: sources,
          pastMessages: pastMessages
        });
        for await (const chunk of result) {
          await writer.write(chunk);
        }
        console.log(result);
      } catch (error) {
        console.error('Streaming error:', error);
      } finally {
        await writer.close();
      }
    })();

    return new Response(stream.readable, {
      headers: {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache, no-transform',
        'Connection': 'keep-alive',
        'X-Accel-Buffering': 'no', 
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET,OPTIONS,PATCH,DELETE,POST,PUT'
      },
    });

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
