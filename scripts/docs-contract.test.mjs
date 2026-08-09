import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import test from "node:test";

const root = new URL("../", import.meta.url).pathname;
const libraryScript = join(root, "scripts/validate-library-selection.mjs");
const extractScript = join(root, "scripts/extract-checked-go.mjs");
const redirectsScript = join(root, "scripts/verify-built-redirects.mjs");
const sha = "5133ec7688f9f12d9ee581fe4311d1524dd2294f";

function run(script, args = [], env = {}) {
  const childEnv = { ...process.env };
  if (Object.hasOwn(env, "GITHUB_OUTPUT")) childEnv.GITHUB_OUTPUT = env.GITHUB_OUTPUT;
  else delete childEnv.GITHUB_OUTPUT;
  return spawnSync(process.execPath, [script, ...args], { cwd: root, encoding: "utf8", env: childEnv });
}

function manifest(value) {
  const dir = mkdtempSync(join(tmpdir(), "docs-manifest-"));
  const path = join(dir, "documented-library.json");
  writeFileSync(path, typeof value === "string" ? value : JSON.stringify(value));
  return path;
}

test("library selection accepts the exact valid manifest", () => {
  const result = run(libraryScript, ["--manifest", manifest({ version: "v0.22.3", sha })]);
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /version=v0\.22\.3\nsha=5133ec/);
});

test("library selection writes step outputs when GitHub provides GITHUB_OUTPUT", () => {
  const output = join(mkdtempSync(join(tmpdir(), "github-output-")), "output");
  const result = run(libraryScript, ["--manifest", manifest({ version: "v0.22.3", sha })], { GITHUB_OUTPUT: output });
  assert.equal(result.status, 0, result.stderr);
  assert.match(readFileSync(output, "utf8"), /version=v0\.22\.3\nsha=5133ec/);
});

for (const [name, value] of [
  ["missing SHA", { version: "v0.22.3" }],
  ["malformed JSON", "{"],
  ["malformed SHA", { version: "v0.22.3", sha: "5133ec7" }],
]) {
  test(`library selection rejects ${name}`, () => {
    const result = run(libraryScript, ["--manifest", manifest(value)]);
    assert.notEqual(result.status, 0);
  });
}

test("dispatch validation rejects a missing SHA", () => {
  const result = run(libraryScript, ["--manifest", manifest({ version: "v0.22.3", sha }), "--dispatch-sha", ""]);
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /missing|invalid/i);
});

test("dispatch validation rejects a mismatched SHA", () => {
  const result = run(libraryScript, ["--manifest", manifest({ version: "v0.22.3", sha }), "--dispatch-sha", "0".repeat(40)]);
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /does not match/);
});

test("dispatch validation accepts the manifest SHA", () => {
  const result = run(libraryScript, ["--manifest", manifest({ version: "v0.22.3", sha }), "--dispatch-sha", sha]);
  assert.equal(result.status, 0, result.stderr);
});

test("extractor selects only explicitly marked complete Go examples", () => {
  const out = mkdtempSync(join(tmpdir(), "checked-go-"));
  const result = run(extractScript, ["--output", out]);
  assert.equal(result.status, 0, result.stderr);
  assert.match(readFileSync(join(out, "first-api-test", "post_api_test.go"), "utf8"), /api\.CallAnApiAt/);
  assert.match(readFileSync(join(out, "notes", "notes_test.go"), "utf8"), /take_notes\.Note\[string\]/);
});

function redirectFixture(target = "/en/get_started/01_installation/", createTarget = true) {
  const dist = mkdtempSync(join(tmpdir(), "redirects-"));
  const sectionDir = join(dist, "en/get_started");
  mkdirSync(sectionDir, { recursive: true });
  writeFileSync(join(sectionDir, "index.html"), `<!doctype html><meta http-equiv="refresh" content="0;url=${target}"><link rel="canonical" href="https://verity-bdd.github.io${target}"><a href="${target}">Redirect</a>`);
  if (createTarget) {
    const targetDir = join(dist, target);
    mkdirSync(targetDir, { recursive: true });
    writeFileSync(join(targetDir, "index.html"), "target");
  }
  return dist;
}

test("built redirect validation checks semantic redirect fields and target artifact", () => {
  const result = run(redirectsScript, ["--dist", redirectFixture(), "--only", "/en/get_started/"]);
  assert.equal(result.status, 0, result.stderr);
});

test("built redirect validation rejects a nonexistent target", () => {
  const target = "/en/get_started/does-not-exist/";
  const result = run(redirectsScript, ["--dist", redirectFixture(target, false), "--only", "/en/get_started/"]);
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /target artifact/i);
});
