#!/usr/bin/env bash
# sync-docs.sh — копирует user-facing docs из lib-репы в Astro content
# Использование: ./scripts/sync-docs.sh <path-to-lib-repo>
# Пример: ./scripts/sync-docs.sh ../lib  (локально)
#          ./scripts/sync-docs.sh ./lib-source  (в CI после checkout)

set -euo pipefail

LIB_DIR="${1:?Usage: sync-docs.sh <path-to-lib-repo>}"
DOCS_CONTENT="src/content/docs/en"

echo "Syncing docs from: $LIB_DIR"
echo "Target: $DOCS_CONTENT"

# --- Markdown guides & examples ---
mkdir -p "$DOCS_CONTENT/guides" "$DOCS_CONTENT/examples"

rsync -av --delete \
  "$LIB_DIR/docs/guides/" \
  "$DOCS_CONTENT/guides/"

rsync -av --delete \
  "$LIB_DIR/docs/examples/" \
  "$DOCS_CONTENT/examples/"

echo "Markdown sync done."
echo "Next step: run gomarkdoc to generate API reference."
