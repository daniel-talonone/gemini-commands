#!/usr/bin/env bash
set -e

# This script handles the deterministic task of setting up a new release branch.
# It ensures the production branch is up-to-date and creates a new release
# branch with a consistent, timestamped name format.

PRODUCTION_BRANCH="production"

# 1. Check if we are on the production branch, if not, check it out.
current_branch=$(git rev-parse --abbrev-ref HEAD)
if [ "$current_branch" != "$PRODUCTION_BRANCH" ]; then
    echo "Switching to '$PRODUCTION_BRANCH' branch..."
    git checkout "$PRODUCTION_BRANCH"
fi

# 2. Pull the latest changes for the production branch.
echo "Updating '$PRODUCTION_BRANCH' branch..."
git pull origin "$PRODUCTION_BRANCH"

# 3. Generate the new release branch name with a precise timestamp.
release_branch_name="release/$(date +%Y%m%d%H%M)"
echo "Creating new release branch: $release_branch_name"

# 4. Create the new branch.
git checkout -b "$release_branch_name"

# 5. Echo the new branch name so the calling process can capture it.
echo "$release_branch_name"
