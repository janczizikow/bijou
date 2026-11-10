#!/usr/bin/env node
const { spawnSync } = require("child_process");
const path = require("path");
const pkg = require("./package.json");

const platform = process.platform;
const arch = process.arch;

// TODO: pick binaries for right platform and arch
const bijou = path.join(__dirname, "bin", "bijou");
const result = spawnSync(bijou, process.argv.slice(2), {
  stdio: "inherit",
  env: { ...process.env, BIJOU_VERSION: pkg.version },
});

if (result.error) {
  console.error("Failed to start executable:", result.error.message);
  process.exit(1);
}

if (result.status !== null) {
  process.exit(result.status);
} else {
  console.error(`Process terminated by signal: ${result.signal}`);
  process.exit(1);
}
