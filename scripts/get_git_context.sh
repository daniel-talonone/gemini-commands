#!/bin/bash

# Gathers git context for the current repository and outputs it as JSON.
# The diff is base64 encoded to ensure it doesn't break the JSON output.

set -e

# Get the current branch name
current_branch=$(git rev-parse --abbrev-ref HEAD)

# Determine the default branch name (e.g., main, master)
default_branch=$(git remote show origin | grep 'HEAD branch' | cut -d' ' -f5)


# Fetch the latest changes from origin to ensure the diff is up-to-date
git fetch origin

# Get the diff between the current branch and the default branch's tracking branch.
# The diff is base64 encoded to handle special characters safely in JSON.
diff_content=$(git diff "origin/$default_branch...HEAD" -- . ':(exclude)vendor' ':(exclude)node_modules')
encoded_diff=$(echo "$diff_content" | base64)

# Output as JSON
cat <<EOF
{
  "current_branch": "$current_branch",
  "default_branch": "$default_branch",
  "diff": "$encoded_diff"
}
EOF
