#!/bin/bash

# Gathers git context for the current repository and outputs it as JSON.
# The diff is base64 encoded to ensure it doesn't break the JSON output.

set -e

# Get the current branch name
current_branch=$(git rev-parse --abbrev-ref HEAD)

# Determine the default branch name (e.g., main, master)
# This command gets the symbolic-ref for the remote's HEAD and strips the 'refs/remotes/origin/' part.
default_branch=$(git symbolic-ref refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@')


# Fetch the latest changes from origin to ensure the diff is up-to-date
git fetch origin

# Get the diff between the current branch and the default branch's tracking branch.
# The diff is base64 encoded to handle special characters safely in JSON.
diff_content=$(git diff "origin/$default_branch...HEAD")
encoded_diff=$(echo "$diff_content" | base64)

# Output as JSON
cat <<EOF
{
  "current_branch": "$current_branch",
  "default_branch": "$default_branch",
  "diff": "$encoded_diff"
}
EOF
