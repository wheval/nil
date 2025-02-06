// Detects and returns the operating system of the user's browser
export function getOperatingSystem() {
  const userAgent = window.navigator.userAgent.toLowerCase();

  if (userAgent.includes("win")) {
    return "windows";
  }
  if (userAgent.includes("mac")) {
    return "mac";
  }
  if (userAgent.includes("linux")) {
    return "linux";
  }

  return "error";
}
