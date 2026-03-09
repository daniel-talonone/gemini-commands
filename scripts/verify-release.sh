#!/usr/bin/env bash
set -e

# --- Constants ---
PRODUCTION_BRANCH="production"
TEMP_DIR=$(mktemp -d)
# Schedule the cleanup of the temporary directory on script exit
trap 'rm -rf -- "$TEMP_DIR"' EXIT

# --- Helper Functions ---

# Normalizes a patch file for comparison.
# This function extracts only the core diff content, starting from the first 'diff --git' line.
# This removes the commit message header and diffstat summary, which can vary.
normalize_patch() {
    local patch_file="$1"
    sed -n '/^diff --git/,$p' "$patch_file"
}

# --- Main Script ---

ORIGINAL_COMMITS_STR=$1
RELEASE_BRANCH=$(git rev-parse --abbrev-ref HEAD)

echo "Starting release verification for branch '$RELEASE_BRANCH'..."
echo "Temporary directory for patches: $TEMP_DIR"
any_issue_found=false

# 1. Get all commits on the release branch since it forked from production.
release_commits=$(git rev-list "${PRODUCTION_BRANCH}..${RELEASE_BRANCH}")

if [ -z "$release_commits" ]; then
    echo "No new commits found on '$RELEASE_BRANCH' since it forked from '$PRODUCTION_BRANCH'."
    exit 0
fi

# 2. Map release commits to original commits and identify any extra (non-cherry-pick) commits.
# Use parallel indexed arrays for bash v3 compatibility (instead of associative arrays).
release_commit_keys=()
original_commit_values=()
extra_commits=()

for r_commit in $release_commits; do
    # Extract the original commit hash from the cherry-pick message in the commit body.
    original_hash=$(git show -s --format=%b "$r_commit" | sed -n 's/.*(cherry picked from commit \([a-f0-9]\{40\}\)).*/\1/p')
    
    if [ -n "$original_hash" ]; then
        release_commit_keys+=("$r_commit")
        original_commit_values+=("$original_hash")
    else
        extra_commits+=("$r_commit")
    fi
done

# 3. Report any extra (non-cherry-pick) commits found on the release branch.
if [ ${#extra_commits[@]} -gt 0 ]; then
    any_issue_found=true
    echo "--------------------------------------------------"
    echo "ERROR: Found commits on the release branch that are NOT cherry-picks:"
    for commit in "${extra_commits[@]}"; do
        git show -s --oneline "$commit"
    done
    echo "These commits must be investigated as they were not part of the intended cherry-pick list."
    echo "--------------------------------------------------"
fi

# 4. Compare patches for all mapped cherry-picked commits.
echo ""
echo "Comparing patches of cherry-picked commits..."

for i in "${!release_commit_keys[@]}"; do
    r_commit=${release_commit_keys[i]}
    o_commit=${original_commit_values[i]}
    
    echo -n "  - Comparing release commit $(git rev-parse --short "$r_commit") with original $(git rev-parse --short "$o_commit")... "

    # Generate patch files for both original and release commits.
    git format-patch --no-stat --stdout -1 "$o_commit" > "$TEMP_DIR/original.patch"
    git format-patch --no-stat --stdout -1 "$r_commit" > "$TEMP_DIR/release.patch"

    # Normalize the patches to compare only the diff content.
    normalize_patch "$TEMP_DIR/original.patch" > "$TEMP_DIR/original.norm.patch"
    normalize_patch "$TEMP_DIR/release.patch" > "$TEMP_DIR/release.norm.patch"

    # Diff the normalized patches and report any differences.
    if ! diff_output=$(diff -u "$TEMP_DIR/original.norm.patch" "$TEMP_DIR/release.norm.patch"); then
        any_issue_found=true
        echo "DIFFERENCE FOUND"
        echo "--------------------------------------------------"
        echo "WARNING: Changes were introduced to commit $o_commit during cherry-pick."
        echo "The final commit in the release is $r_commit."
        echo "Review this diff of the patches to ensure no unintended logic was added:"
        echo ""
        echo "$diff_output"
        echo "--------------------------------------------------"
    else
        echo "OK"
    fi
done

# 5. Check if all user-provided original commits were actually picked.
IFS=',' read -ra user_original_commits <<< "$ORIGINAL_COMMITS_STR"
missing_commits=()

for user_commit in "${user_original_commits[@]}"; do
    full_user_commit=$(git rev-parse "$user_commit")
    found=false
    # Perform a linear search through the original commit values.
    for o_val in "${original_commit_values[@]}"; do
        if [ "$o_val" == "$full_user_commit" ]; then
            found=true
            break
        fi
    done

    if [ "$found" = false ]; then
        missing_commits+=("$user_commit")
    fi
done

if [ ${#missing_commits[@]} -gt 0 ]; then
    any_issue_found=true
    echo ""
    echo "--------------------------------------------------"
    echo "ERROR: The following commits were expected but NOT found in the release branch:"
    for commit in "${missing_commits[@]}"; do
        echo "  - $commit"
    done
    echo "--------------------------------------------------"
fi

# 6. Final Report
echo ""
echo "=================================================="
if [ "$any_issue_found" = true ]; then
    echo "VERIFICATION FAILED: Issues were found. Please review the output above carefully."
    exit 1
else
    echo "VERIFICATION SUCCESSFUL: The release branch is consistent with the provided original commits."
fi
