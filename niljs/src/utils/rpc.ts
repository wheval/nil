function isValidHttpHeaders(headers: unknown) {
  if (headers === null || typeof headers !== "object" || Array.isArray(headers)) {
    throw new Error("Invalid headers provided to the RPC client.");
  }

  const isValidObj = Object.entries(headers).every(
    ([key, value]) => typeof key === "string" && typeof value === "string",
  );

  if (!isValidObj) {
    throw new Error("Invalid http headers provided to the RPC client.");
  }
}

export { isValidHttpHeaders };
