export async function focusOrCreateWelcomeTab(): Promise<void> {
  const extension = await chrome.management.getSelf();

  const tabs = await chrome.tabs.query({
    url: `chrome-extension://${extension.id}/welcome.html*`,
  });
  const tab = tabs[0];

  const url = "welcome.html#";

  if (!tab?.id) {
    await chrome.tabs.create({ url });
    return;
  }

  // Focus the existing tab if it's already open
  await chrome.tabs.update(tab.id, { active: true });
}
