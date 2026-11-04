#!/usr/bin/env bash

set -euo pipefail

# Decide whether a push needs the lint and test jobs, and print `code=<bool>` for ci.yml to append
# to $GITHUB_OUTPUT.
#
# Usage: detect-changes.sh <before-sha>
#
# On master the diff starts at <before-sha>, the tip before this push. Anywhere else it starts at
# the merge base with origin/master, so a branch that changed code in any commit keeps running
# every job after a docs-only follow-up. code is false only when every changed file is
# documentation. Manual runs, any doubt about the base, or an empty diff run everything.

readonly before="${1:-}"

readonly DOCS_PATTERN='^([^/]+\.md|\.github/ISSUE_TEMPLATE/.*|\.github/pull_request_template\.md|CODEOWNERS|LICENSE\.txt)$'

function main {
	local base changed

	if [[ "${GITHUB_EVENT_NAME:-}" == "workflow_dispatch" ]]; then
		echo "Manual run; running every job" >&2
		echo "code=true"
		return 0
	fi

	base="$(resolve_base)"

	if [[ -z "$base" ]]; then
		echo "Could not resolve a base commit; running every job" >&2
		echo "code=true"
		return 0
	fi

	changed="$(git diff --name-only "$base" HEAD)"

	if [[ -z "$changed" ]]; then
		echo "No changes against ${base}; running every job" >&2
		echo "code=true"
		return 0
	fi

	if grep -qvE "$DOCS_PATTERN" <<<"$changed"; then
		echo "code=true"
		return 0
	fi

	echo "Only documentation changed against ${base}; skipping lint and tests" >&2
	echo "code=false"

	return 0
}

# Print the commit the diff starts from, or nothing when it cannot be trusted.
function resolve_base {
	if [[ "${GITHUB_REF:-}" != "refs/heads/master" ]]; then
		git merge-base origin/master HEAD 2>/dev/null || true
		return 0
	fi

	# A branch creation or a force push leaves no usable previous tip.
	if [[ -z "$before" || "$before" =~ ^0+$ ]]; then
		return 0
	fi

	if ! git merge-base --is-ancestor "$before" HEAD 2>/dev/null; then
		return 0
	fi

	echo "$before"

	return 0
}

main
