"use client";
import { Thread } from "@assistant-ui/react";
import { MarkdownText } from "../MarkdownText";



export function AIChat() {
  return (
    <>
      <Thread
        assistantAvatar={{ src: "/img/nil-logo.png" }}
        assistantMessage={{ components: { Text: MarkdownText } }}
        composer={{ allowAttachments: false }}
        welcome={{ message: "Got questions about =nil;, zkSharding, smart contracts or anything else? Just send me a message!" }}
      />
    </>
  );
}
