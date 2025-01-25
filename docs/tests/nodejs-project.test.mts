import { type ChildProcess, spawn } from "node:child_process";
import fetch from "node-fetch";
import { NEW_SMART_ACCOUNT_PATTERN, SERVER_RUNNING_PATTERN } from "./patterns";

const PATH_TO_PROJECT = "./tests/nodeJsProject/server.js";

describe("the simple Node.js project correctly creates a smart account", () => {
  let serverProcess: ChildProcess;

  beforeAll((done) => {
    serverProcess = spawn(process.env.NODE_JS, [PATH_TO_PROJECT], { stdio: "pipe" });

    serverProcess.stdout.setEncoding("utf8");
    serverProcess.stderr.setEncoding("utf8");

    let serverOutput = "";

    serverProcess.stdout.on("data", (data) => {
      console.log(`stdout: ${data}`);
      serverOutput += data;
      expect(serverOutput).toMatch(SERVER_RUNNING_PATTERN);
    });

    serverProcess.stderr.on("data", (data) => {
      console.error(`stderr: ${data}`);
    });

    serverProcess.on("error", (err) => {
      console.error(`Server process error: ${err}`);
    });

    serverProcess.on("close", (code) => {
      console.log(`Server process exited with code ${code}`);
    });
  });

  afterAll(() => {
    if (serverProcess) {
      serverProcess.kill();
    }
  });

  test("the Node.js project creates a smart account", async () => {
    await new Promise((resolve) => setTimeout(resolve, 7000));

    try {
      const response = await fetch("http://127.0.0.1:3000/");
      expect(response.ok).toBe(true);
      const text = await response.text();
      expect(text).toMatch(NEW_SMART_ACCOUNT_PATTERN);
    } catch (error) {
      console.error("Fetch error:", error);
      throw error;
    }
  });
});
