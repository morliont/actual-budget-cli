#!/usr/bin/env bash
set -euo pipefail

current_tag="${1:-${GITHUB_REF_NAME:-}}"
if [[ -z "${current_tag}" ]]; then
  current_tag="$(git describe --tags --abbrev=0 2>/dev/null || true)"
fi

prev_tag=""
if [[ -n "${current_tag}" ]]; then
  prev_tag="$(git tag --sort=-creatordate | awk -v tag="$current_tag" '$0 != tag {print; exit}')"
fi

if [[ -n "${prev_tag}" ]]; then
  range="${prev_tag}..HEAD"
else
  range="HEAD"
fi

printf "## What's Changed\n\n"
printf "Generated from commit history (%s).\n\n" "$range"

git log --reverse --pretty=format:'- %s (%h)' "$range" || true

printf "\n\n## Full Changelog\n\n"
if [[ -n "${prev_tag}" && -n "${current_tag}" ]]; then
  printf -- "- %s...%s\n" "$prev_tag" "$current_tag"
else
  printf -- "- Initial release\n"
fi
