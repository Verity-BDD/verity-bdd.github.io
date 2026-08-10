#!/usr/bin/env node
import { mkdirSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import { basename, join } from "node:path";

function fail(message) {
  console.error(`extract-checked-go: ${message}`);
  process.exit(1);
}

const args = process.argv.slice(2);
if (args.length !== 2 || args[0] !== "--output") fail("usage: extract-checked-go.mjs --output DIR");
const output = args[1];
const docs = new URL("../src/content/docs/en/", import.meta.url).pathname;
const expected = new Set(["first-api-test/post_api_test.go", "notes/notes_test.go"]);
const found = new Set();

function markdownFiles(dir) {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    return entry.isDirectory() ? markdownFiles(path) : entry.name.endsWith(".md") ? [path] : [];
  });
}

const marker = /<!-- checked-go example=([a-z0-9-]+) file=([A-Za-z0-9_]+\.go) -->\s*```go(?:[^\n]*)\n([\s\S]*?)\n```\s*<!-- \/checked-go -->/g;
for (const path of markdownFiles(docs)) {
  const body = readFileSync(path, "utf8");
  for (const match of body.matchAll(marker)) {
    const key = `${match[1]}/${match[2]}`;
    if (!expected.has(key)) fail(`unexpected checked example ${key} in ${basename(path)}`);
    if (found.has(key)) fail(`duplicate checked example ${key}`);
    if (!/^package\s+[A-Za-z_][A-Za-z0-9_]*/m.test(match[3])) fail(`${key} is not a complete Go source file`);
    const dir = join(output, match[1]);
    mkdirSync(dir, { recursive: true });
    writeFileSync(join(dir, match[2]), `${match[3]}\n`);
    found.add(key);
  }
}
for (const key of expected) if (!found.has(key)) fail(`missing checked example marker ${key}`);
console.log(`extracted ${found.size} checked Markdown Go examples`);
