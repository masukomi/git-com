#!/usr/bin/env bash

set -eo pipefail

function remote_has_tag () {
    git ls-remote --tags $1 2>/dev/null | sed -e 's/.*refs\/tags\///; s/\^{}$//' | grep --color=none $2 > /dev/null
}



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

# BEGIN CONFIRM GITHUB RELEASE
gh release list | grep "$TAG" > /dev/null

if [[ $? -ne 0 ]]; then
    echo "There's no release for tag $TAG on GitHub"
    # 🤔 is there a tag?
    GH_REMOTE=$(git remote -vv | grep --color=none "github\.com" | sed '/(fetch)/d'| awk '{print $1}')
    if [ remote_has_tag $GH_REMOTE $TAG ]; then
        echo "GitHub has that tag"
    else
        # sanity check we even have the tag locally
        git tag | grep --color=none "$TAG"
        if [ $? -ne 0 ]; then
            echo "Something's wrong. Can't find tag \"$TAG\" locally."
            exit 64 # EX_USAGE
        fi
        echo "GitHub does NOT have the $TAG tag"
        echo "Will push."
        git push --tags $GH_REMOTE
    fi

    echo "Creating release for that tag on GitHub"
    gh release create "$TAG" --notes-from-tag
else
    echo "GitHub has a release for $TAG"
fi
# DONE CONFIRMING GITHUB RELEASE

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
