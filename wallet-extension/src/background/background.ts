import { focusOrCreateWelcomeTab } from "./util.ts";

chrome.runtime.onInstalled.addListener(async ({ reason }) => {
  if (reason === "install") {
    // Open welcome page when the extension is first installed
    await focusOrCreateWelcomeTab();
  }
});

chrome.commands.onCommand.addListener(async (command) => {
  if (command === "open-popup") {
    // Opens the extension popup
    await chrome.action.openPopup();
  }
});
