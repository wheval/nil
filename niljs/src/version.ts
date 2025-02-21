import pkgJson from "../package.json" with { type: "json" };

const version = pkgJson.version;

export { version };
