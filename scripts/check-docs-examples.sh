#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
lib_dir=${1:?usage: scripts/check-docs-examples.sh LIBRARY_CHECKOUT}
version=$(node -p "require('$repo_root/documented-library.json').version")
sha=$(node -p "require('$repo_root/documented-library.json').sha")

actual_sha=$(git -C "$lib_dir" rev-parse HEAD)
[[ "$actual_sha" == "$sha" ]] || { echo "library HEAD $actual_sha does not match documented SHA $sha" >&2; exit 1; }
[[ -z $(git -C "$lib_dir" status --porcelain --untracked-files=normal) ]] || { echo "library checkout is dirty" >&2; exit 1; }
grep -Fxq "require github.com/verity-bdd/verity-bdd $version" "$repo_root/scripts/checked-examples/go.mod" || {
  echo "checked examples do not require documented version $version" >&2
  exit 1
}

if grep -RInE 'api\.(NewResponseHeader|NewJSONPath|NewRequestBuilder)\(' "$repo_root/src/content/docs/en"; then
  echo "hand-written docs contain removed v0.22.3 API calls" >&2
  exit 1
fi
if grep -RInPzo 'WithJSONBody\([^\n]*\)\s*\.' "$repo_root/src/content/docs/en"; then
  echo "WithJSONBody returns error and cannot be chained" >&2
  exit 1
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
cp -R "$repo_root/scripts/checked-examples/." "$tmp/"
(
  cd "$tmp"
  go mod edit -replace "github.com/verity-bdd/verity-bdd=$lib_dir"
  go mod tidy
  go test ./...
)
printf 'checked Go examples compile against Verity BDD %s (%s)\n' "$version" "$sha"
