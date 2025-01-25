import type { ReactNode } from "react";
import {
  AssistantRuntimeProvider,
  useLocalRuntime,
  type ChatModelAdapter,
} from "@assistant-ui/react";
import { useCallback, useEffect, useMemo, useState } from "react";

import { useGoogleReCaptcha } from "react-google-recaptcha-v3";

const pattern = /\d+:"([^"]*)"/g;

function asAsyncIterable<T>(source: ReadableStream<T>): AsyncIterable<T> {
  return {
    [Symbol.asyncIterator]: () => {
      const reader = source.getReader();
      return {
        async next(): Promise<IteratorResult<T, undefined>> {
          const { done, value } = await reader.read();
          return done ? { done: true, value: undefined } : { done: false, value };
        },
      };
    },
  };
}

const CustomModelAdapter: (string, Function) => ChatModelAdapter = (
  token,
  handleReCaptchaVerify,
) => ({
  async *run({ messages, abortSignal }) {
    const messagesToSend = messages.map((m) => ({
      role: m.role,
      content: m.content
        .filter((c) => c.type === "text")
        .map((c) => c.text)
        .join(" "),
    }));
    const response = await fetch("https://docs.nil.foundation/bot/api/chat", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        messages: messagesToSend,
        token: token,
      }),
      signal: abortSignal,
    });
    handleReCaptchaVerify();

    let text = "";
    for await (const chunk of asAsyncIterable(
      response.body!.pipeThrough(new TextDecoderStream()),
    )) {
      const matches = [...chunk.matchAll(pattern)];
      const tokens = matches.map((match) => match[1]);

      let cleanedText = tokens.join("");
      cleanedText = cleanedText.replace(/\\n/g, "\n");
      text += cleanedText;
      yield { content: [{ type: "text", text }] };
    }
  },
});

export function CustomRuntimeProvider({
  children,
}: Readonly<{
  children: ReactNode;
}>) {
  const [token, tokenSetter] = useState<string | null>(null);

  const { executeRecaptcha } = useGoogleReCaptcha();

  const handleReCaptchaVerify = useCallback(async () => {
    if (!executeRecaptcha) {
      return;
    }

    const t = await executeRecaptcha();
    tokenSetter(t);
  }, [executeRecaptcha]);

  useEffect(() => {
    handleReCaptchaVerify();
  }, [handleReCaptchaVerify]);

  const adapter = useMemo(() => {
    return CustomModelAdapter(token, handleReCaptchaVerify);
  }, [token]);
  const runtime = useLocalRuntime(adapter);

  return <AssistantRuntimeProvider runtime={runtime}>{children}</AssistantRuntimeProvider>;
}
