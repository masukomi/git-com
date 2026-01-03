#!/usr/bin/env bash
set -eo pipefail

if [ -z "$FORMULA_PATH" ]; then
    FORMULA_PATH="$HOME/workspace/homebrew-tap/Formula/git-com.rb"
fi
REPO_URL="https://github.com/masukomi/git-com"

usage() {
    echo "Usage: $0 <tag>"
    echo "Example: $0 v1.1.0"
    exit 1
}

if [[ $# -ne 1 ]]; then
    usage
fi

TAG="$1"

# Validate tag format (should start with 'v')
if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Tag must be in format v#.#.# (e.g., v1.1.0)"
    exit 1
fi

# Extract version number (strip the 'v' prefix)
VERSION="${TAG#v}"

# Check that formula file exists
if [[ ! -f "$FORMULA_PATH" ]]; then
    echo "Error: Formula not found at $FORMULA_PATH"
    exit 1
fi

TARBALL_URL="${REPO_URL}/archive/refs/tags/${TAG}.tar.gz"

echo "Fetching tarball for ${TAG}..."
SHA256=$(curl -sL "$TARBALL_URL" | shasum -a 256 | awk '{print $1}')

if [[ -z "$SHA256" || ${#SHA256} -ne 64 ]]; then
    echo "Error: Failed to calculate SHA256. Does the tag ${TAG} exist?"
    exit 1
fi

echo "SHA256: ${SHA256}"

# Update the formula
sed -i '' "s/current_version=\"[^\"]*\"/current_version=\"${VERSION}\"/" "$FORMULA_PATH"
sed -i '' "s/sha256 \"[^\"]*\"/sha256 \"${SHA256}\"/" "$FORMULA_PATH"

echo "Updated ${FORMULA_PATH} to version ${VERSION}"
