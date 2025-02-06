import { NextResponse } from 'next/server';
import axios from "axios";
import hash from 'hash-it';
import clientInstance from '@/clients/DBClient';

const secretKey = process.env.RECAPTCHA_SECRET_KEY;

const headers = {
  'Content-Type': 'application/json',
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET,OPTIONS,PATCH,DELETE,POST,PUT'
};

const checkCaptcha = async (token: string) => {
  const result = await axios.post(
    `https://www.google.com/recaptcha/api/siteverify?secret=${secretKey}&response=${token}`
  );

  return result.data.success;
}

export async function POST(req: Request) {
  const data = await req.json();
  const { rpc, token } = data;

  if (!token) {
    return new Response(JSON.stringify({ message: "Token not found" }), {
      status: 405,
    });
  }

  try {
    const captchaResult = await checkCaptcha(token);

    if (captchaResult) {
      const rpcAddress = rpc.at(-1).content;
      const rpcHash = hash(rpcAddress).toString();

      await clientInstance.insertHash(rpcHash);

      return new Response(null, {
        status: 200,
        headers: headers,
      });
    } else {
      return new Response(JSON.stringify({ message: "Failed to verify" }), {
        status: 405,
        headers: headers
      });
    }
  } catch (error) {
    console.log(error);
    return new Response(JSON.stringify({ message: "Internal server error" }), {
      status: 500,
      headers: headers
    });
  }
}

export async function PUT(req: Request) {
  const data = await req.json();
  const { rpc, progress, token } = data;

  if (!token) {
    return new Response(JSON.stringify({ message: "Token not found" }), {
      status: 405,
    });
  }

  try {
    const captchaResult = await checkCaptcha(token);

    if (captchaResult) {
      const rpcAddress = rpc.at(-1).content;
      const newStage = progress.at(-1).content;
      const rpcHash = hash(rpcAddress).toString();

      await clientInstance.updateProgress(rpcHash, newStage);

      return new Response(null, {
        status: 200,
        headers: headers,
      });
    } else {
      return new Response(JSON.stringify({ message: "Failed to verify" }), {
        status: 405,
        headers: headers
      });
    }
  } catch (error) {
    console.log(error);
    return new Response(JSON.stringify({ message: "Internal server error" }), {
      status: 500,
      headers: headers
    });
  }
}

export async function GET(req: Request) {
  const data = await req.json();
  const { rpc, token } = data;


  if (!token) {
    return new Response(JSON.stringify({ message: "Token not found" }), {
      status: 405,
    });
  }

  try {
    const captchaResult = await checkCaptcha(token);

    if (captchaResult) {
      const rpcAddress = rpc.at(-1).content;
      const rpcHash = hash(rpcAddress).toString();

      const result = await clientInstance.retrieveProgress(rpcHash);

      return new Response(JSON.stringify({ result: result }), {
        status: 200,
        headers: headers,
      });
    } else {
      return new Response(JSON.stringify({ message: "Failed to verify" }), {
        status: 405,
        headers: headers
      });
    }
  } catch (error) {
    console.log(error);
    return new Response(JSON.stringify({ message: "Internal server error" }), {
      status: 500,
      headers: headers
    });
  }
}

export async function DELETE(req: Request) {
  const data = await req.json();
  const { rpc, token } = data;


  if (!token) {
    return new Response(JSON.stringify({ message: "Token not found" }), {
      status: 405,
    });
  }

  try {
    const captchaResult = await checkCaptcha(token);

    if (captchaResult) {
      const rpcAddress = rpc.at(-1).content;
      const rpcHash = hash(rpcAddress).toString();

      const result = await clientInstance.removeUser(rpcHash);

      return new Response(JSON.stringify({ result: result }), {
        status: 200,
        headers: headers,
      });
    } else {
      return new Response(JSON.stringify({ message: "Failed to verify" }), {
        status: 405,
        headers: headers
      });
    }
  } catch (error) {
    console.log(error);
    return new Response(JSON.stringify({ message: "Internal server error" }), {
      status: 500,
      headers: headers
    });
  }
}

export async function OPTIONS() {
  return new NextResponse(null, {
    status: 200,
    headers: headers
  })
}