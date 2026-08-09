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
  const attributes = tag.replace(/^<\s*[^\s/>]+/, "");
  const pattern = new RegExp(`(?:^|\\s)${name}\\s*=\\s*(?:"([^"]*)"|'([^']*)'|([^\\s"'=<>\u0060]+))`, "gi");
  const matches = [...attributes.matchAll(pattern)];
  return matches.length === 1 ? matches[0][1] ?? matches[0][2] ?? matches[0][3] : undefined;
}
function hasAttribute(tag, name) {
  const attributes = tag.replace(/^<\s*[^\s/>]+/, "");
  return new RegExp(`(?:^|\\s)${name}(?=\\s|=|/?>)`, "i").test(attributes);
}

const inertElements = new Set(["iframe", "noembed", "noframes", "plaintext", "script", "style", "textarea", "title", "xmp"]);

function tagEnd(html, start) {
  let quote;
  for (let index = start + 1; index < html.length; index += 1) {
    const character = html[index];
    if (quote) {
      if (character === quote) quote = undefined;
    } else if (character === '"' || character === "'") quote = character;
    else if (character === ">") return index;
  }
  return -1;
}

function activeStartTags(html) {
  const tags = [];
  let index = 0;
  let rawTextElement;
  let templateDepth = 0;

  while (index < html.length) {
    if (rawTextElement) {
      if (rawTextElement === "plaintext") break;
      const closing = new RegExp(`</\\s*${rawTextElement}(?=[\\s>/])`, "ig");
      closing.lastIndex = index;
      const match = closing.exec(html);
      if (!match) break;
      const end = tagEnd(html, match.index);
      if (end < 0) break;
      rawTextElement = undefined;
      index = end + 1;
      continue;
    }

    const start = html.indexOf("<", index);
    if (start < 0) break;
    if (html.startsWith("<!--", start)) {
      const end = html.indexOf("-->", start + 4);
      if (end < 0) break;
      index = end + 3;
      continue;
    }

    const end = tagEnd(html, start);
    if (end < 0) break;
    const tag = html.slice(start, end + 1);
    const closing = tag.match(/^<\s*\/\s*([a-z][\w:-]*)/i);
    if (closing) {
      if (closing[1].toLowerCase() === "template" && templateDepth > 0) templateDepth -= 1;
      index = end + 1;
      continue;
    }

    const opening = tag.match(/^<\s*([a-z][\w:-]*)/i);
    if (!opening) {
      index = end + 1;
      continue;
    }
    const name = opening[1].toLowerCase();
    if (name === "template") templateDepth += 1;
    else {
      if (templateDepth === 0) tags.push({ name, tag });
      if (inertElements.has(name)) rawTextElement = name;
    }
    index = end + 1;
  }
  return tags;
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
  const tags = activeStartTags(readFileSync(artifact, "utf8"));
  const refreshTags = tags.filter(({ name, tag }) => name === "meta" && hasAttribute(tag, "http-equiv"));
  if (refreshTags.length !== 1) fail(`${route} must contain exactly one refresh element; found ${refreshTags.length}`);
  if (refreshTags.length === 1 && attribute(refreshTags[0].tag, "http-equiv")?.trim().toLowerCase() !== "refresh") {
    fail(`${route} http-equiv is not refresh`);
  }
  const refresh = refreshTags.length === 1 && attribute(refreshTags[0].tag, "content")?.match(/^\s*\d+\s*;\s*url=(.+)\s*$/i)?.[1];
  if (refresh !== target) fail(`${route} refresh target ${JSON.stringify(refresh)} != ${target}`);
  const canonicalTags = tags.filter(({ name }) => name === "link");
  if (canonicalTags.length !== 1) fail(`${route} must contain exactly one canonical element; found ${canonicalTags.length}`);
  if (canonicalTags.length === 1 && !attribute(canonicalTags[0].tag, "rel")?.toLowerCase().split(/\s+/).includes("canonical")) {
    fail(`${route} link is not canonical`);
  }
  const canonical = canonicalTags.length === 1 && attribute(canonicalTags[0].tag, "href");
  if (canonical !== site + target) fail(`${route} canonical URL ${JSON.stringify(canonical)} != ${site + target}`);
  const fallbackTags = tags.filter(({ name }) => name === "a");
  if (fallbackTags.length !== 1) fail(`${route} must contain exactly one fallback link; found ${fallbackTags.length}`);
  const fallback = fallbackTags.length === 1 && attribute(fallbackTags[0].tag, "href");
  if (fallback !== target) fail(`${route} fallback target ${JSON.stringify(fallback)} != ${target}`);
  const targetArtifact = join(dist, target.slice(1), "index.html");
  if (!existsSync(targetArtifact)) fail(`${route} target artifact does not exist: ${targetArtifact}`);
}
if (!process.exitCode) console.log(`verified ${selected.length} built section redirects`);
