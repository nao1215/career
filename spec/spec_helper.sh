#!/bin/sh
# shellcheck shell=sh
#
# shellspec helper for career end-to-end tests. These drive the binary built by
# `make build` the way a user does (subcommands, flags, exit codes, files on
# disk) so they catch regressions the Go tests cannot.

set -eu

PROJECT_ROOT="$(cd "$SHELLSPEC_SPECDIR/.." && pwd)"
export PROJECT_ROOT

# CAREER_BIN points at the binary built by `make build`. Override to test another
# build.
CAREER_BIN="${CAREER_BIN:-$PROJECT_ROOT/career}"
export CAREER_BIN

# EXAMPLES is the bundled sample directory.
EXAMPLES="$PROJECT_ROOT/examples"
export EXAMPLES

# career runs the built binary.
career() {
  "$CAREER_BIN" "$@"
}

# pdf_magic prints the first four bytes of a file so specs can assert it is a
# PDF without depending on external tools.
pdf_magic() {
  head -c 4 "$1"
}

make_workdir() {
  WORK="$(mktemp -d)"
  cp "$EXAMPLES/minimal.yaml" "$WORK/resume.yaml"
  export WORK
}

remove_workdir() {
  if [ -n "${WORK:-}" ]; then
    rm -rf "$WORK"
  fi
}
