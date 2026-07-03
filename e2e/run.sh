#!/usr/bin/env bash
#
# run.sh builds career from this checkout and runs the atago end-to-end suite
# (e2e/atago/*.atago.yaml) against the real binary.
#
# The test DEFINITIONS are atago YAML — this script is only the environment
# bootstrap (a plain shell program, not a test framework). Each scenario works
# inside its isolated ${workdir}; specs that need a bundled sample resume copy
# it from $CAREER_EXAMPLES.
#
# Environment contract used by the specs:
#   PATH              career resolves here (built from this checkout)
#   CAREER_EXAMPLES   absolute path to the bundled examples/ fixtures
#
# Usage: e2e/run.sh [atago args...]        (e.g. e2e/run.sh --filter generate)
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"

if ! command -v atago >/dev/null 2>&1; then
	echo "e2e: atago is not installed. Install it from https://github.com/nao1215/atago" >&2
	echo "e2e: e.g. 'go install github.com/nao1215/atago@latest' (CI uses nao1215/setup-atago)" >&2
	exit 127
fi

TMP="$(mktemp -d "${TMPDIR:-/tmp}/career-e2e.XXXXXX")"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT
mkdir -p "$TMP/bin"

echo "e2e: building career..."
(cd "$REPO_ROOT" && env CGO_ENABLED=0 go build -o "$TMP/bin/career" .)

# Put the e2e-built career first on PATH so the specs exercise that binary.
export PATH="$TMP/bin:$PATH"
export CAREER_EXAMPLES="$REPO_ROOT/examples"

echo "e2e: career $("$TMP/bin/career" version | head -1)"
# Extra args (e.g. --filter X) go before the path so the flag parser sees them.
atago run "$@" "$SCRIPT_DIR/atago"
