#!/usr/bin/env node
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";

const redirects = new Map([
  ["/en/get_started/", "/en/get_started/01_installation/"],
  ["/en/core_concepts/", "/en/core_concepts/1_screenplay/"],
  ["/en/guides/", "/en/guides/11_notes/"],
  ["/en/examples/", "/en/examples/abilities/"],
  ["/en/api/", "/en/api/verity-bdd/"],
]);
const site = "https://verity-bdd.github.io";

function fail(message) {
  console.error(`built-redirects: ${message}`);
  process.exitCode = 1;
}
function attribute(tag, name) {
  return tag.match(new RegExp(`\\b${name}=["']([^"']+)["']`, "i"))?.[1];
}

const args = process.argv.slice(2);
let dist;
let only;
for (let index = 0; index < args.length; index += 2) {
  if (args[index] === "--dist") dist = args[index + 1];
  else if (args[index] === "--only") only = args[index + 1];
  else fail(`unknown argument ${args[index]}`);
}
if (!dist) fail("--dist is required");
const selected = only ? [[only, redirects.get(only)]] : [...redirects];
if (only && !redirects.has(only)) fail(`unknown section route ${only}`);

for (const [route, target] of selected) {
  if (!target) continue;
  const artifact = join(dist, route.slice(1), "index.html");
  if (!existsSync(artifact)) {
    fail(`missing section artifact ${artifact}`);
    continue;
  }
  const html = readFileSync(artifact, "utf8");
  const meta = html.match(/<meta\b[^>]*>/gi)?.find((tag) => attribute(tag, "http-equiv")?.toLowerCase() === "refresh");
  const refresh = meta && attribute(meta, "content")?.match(/^\s*\d+\s*;\s*url=(.+)\s*$/i)?.[1];
  if (refresh !== target) fail(`${route} refresh target ${JSON.stringify(refresh)} != ${target}`);
  const canonicalTag = html.match(/<link\b[^>]*>/gi)?.find((tag) => attribute(tag, "rel")?.toLowerCase() === "canonical");
  const canonical = canonicalTag && attribute(canonicalTag, "href");
  if (canonical !== site + target) fail(`${route} canonical URL ${JSON.stringify(canonical)} != ${site + target}`);
  const fallback = html.match(/<a\b[^>]*>/gi)?.some((tag) => attribute(tag, "href") === target);
  if (!fallback) fail(`${route} is missing fallback link ${target}`);
  const targetArtifact = join(dist, target.slice(1), "index.html");
  if (!existsSync(targetArtifact)) fail(`${route} target artifact does not exist: ${targetArtifact}`);
}
if (!process.exitCode) console.log(`verified ${selected.length} built section redirects`);
