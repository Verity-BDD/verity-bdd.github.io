#!/usr/bin/env node
import { appendFileSync, readFileSync } from "node:fs";

function fail(message) {
  console.error(`library-selection: ${message}`);
  process.exit(1);
}

const args = process.argv.slice(2);
let manifestPath;
let dispatchSHA;
let dispatchRequested = false;
for (let index = 0; index < args.length; index += 2) {
  const flag = args[index];
  const value = args[index + 1];
  if (value === undefined) fail(`missing value for ${flag}`);
  if (flag === "--manifest") manifestPath = value;
  else if (flag === "--dispatch-sha") {
    dispatchRequested = true;
    dispatchSHA = value;
  } else fail(`unknown argument ${flag}`);
}
if (!manifestPath) fail("--manifest is required");

let manifest;
try {
  manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
} catch (error) {
  fail(`cannot read valid manifest ${manifestPath}: ${error.message}`);
}
if (!manifest || Array.isArray(manifest) || typeof manifest !== "object") fail("manifest must be a JSON object");
if (typeof manifest.version !== "string" || !/^v[0-9]+\.[0-9]+\.[0-9]+$/.test(manifest.version)) fail("invalid or missing documented version");
if (typeof manifest.sha !== "string" || !/^[0-9a-f]{40}$/.test(manifest.sha)) fail("invalid or missing documented SHA");
if (dispatchRequested) {
  if (!/^[0-9a-f]{40}$/.test(dispatchSHA)) fail("invalid or missing repository_dispatch lowercase library SHA");
  if (dispatchSHA !== manifest.sha) fail(`dispatch SHA ${dispatchSHA} does not match documented SHA ${manifest.sha}`);
}

const output = `version=${manifest.version}\nsha=${manifest.sha}\n`;
if (process.env.GITHUB_OUTPUT) appendFileSync(process.env.GITHUB_OUTPUT, output);
else process.stdout.write(output);
