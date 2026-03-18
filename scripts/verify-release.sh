#!/usr/bin/env bash
set -e

# --- Constants ---
PRODUCTION_BRANCH="production"
TEMP_DIR=$(mktemp -d)
# Schedule the cleanup of the temporary directory on script exit
trap 'rm -rf -- "$TEMP_DIR"' EXIT

# --- Helper Functions ---
# Normalizes a patch file to extract only the core diff content.
normalize_patch() {
    sed -n '/^diff --git/,$p' "$1"
}

# Escapes a string for use in JSON.
json_escape() {
    # Using python to robustly escape the string for JSON
    python3 -c 'import json, sys; print(json.dumps(sys.stdin.read()))'
}

# --- Main Script ---

ORIGINAL_COMMITS_STR=$1
RELEASE_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# 1. Get all commits on the release branch since it forked from production.
# Use a while-read loop for portability instead of mapfile.
declare -a release_commits=()
while IFS= read -r line; do
    # Avoid adding empty lines to the array if git rev-list is empty
    if [ -n "$line" ]; then
        release_commits+=("$line")
    fi
done < <(git rev-list "${PRODUCTION_BRANCH}..${RELEASE_BRANCH}")

# If there are no new commits, output an empty success JSON and exit.
if [ ${#release_commits[@]} -eq 0 ]; then
    echo '{"status": "VERIFICATION_SUCCESSFUL", "extra_commits": [], "missing_commits": [], "changed_commits": []}'
    exit 0
fi

# 2. Parse user-provided original commits into an array.
declare -a user_original_commits=()
IFS=',' read -ra user_original_commits_raw <<< "$ORIGINAL_COMMITS_STR"
for commit in "${user_original_commits_raw[@]}"; do
    if [ -n "$commit" ]; then
        user_original_commits+=("$(git rev-parse "$commit")")
    fi
done

# 3. Initialize arrays for findings.
declare -a mapped_release_commits=()
declare -a mapped_original_commits=()
declare -a unmapped_release_commits=("${release_commits[@]}")
declare -a changed_commits_json_parts=()

# 4. Map commits using the reliable '(cherry picked...)' string first.
declare -a found_indices=()
for i in "${!unmapped_release_commits[@]}"; do
    r_commit=${unmapped_release_commits[i]}
    original_hash=$(git show -s --format=%b "$r_commit" | sed -n 's/.*(cherry picked from commit \([a-f0-9]\{40\}\)).*/\1/p')
    if [ -n "$original_hash" ]; then
        mapped_release_commits+=("$r_commit")
        mapped_original_commits+=("$original_hash")
        found_indices+=($i)
    fi
done
# Remove found commits from unmapped list (in reverse to not mess up indices)
for ((i=${#found_indices[@]}-1; i>=0; i--)); do
    unset "unmapped_release_commits[${found_indices[i]}]"
done
unmapped_release_commits=("${unmapped_release_commits[@]}") # Re-index array

# 5. Fallback: Map remaining commits by searching for matching subject lines.
declare -a temp_user_commits=("${user_original_commits[@]}")
declare -a newly_mapped_indices=()
for i in "${!unmapped_release_commits[@]}"; do
    r_commit=${unmapped_release_commits[i]}
    r_commit_subject=$(git show -s --format=%s "$r_commit")
    
    for j in "${!temp_user_commits[@]}"; do
        o_commit=${temp_user_commits[j]}
        o_commit_subject=$(git show -s --format=%s "$o_commit")

        if [ "$r_commit_subject" == "$o_commit_subject" ]; then
            # Found a match by subject
            mapped_release_commits+=("$r_commit")
            mapped_original_commits+=("$o_commit")
            newly_mapped_indices+=($i)
            # Remove from temp lists to prevent re-matching
            unset "temp_user_commits[j]"
            break
        fi
    done
done
# Remove newly found commits from unmapped list
for ((i=${#newly_mapped_indices[@]}-1; i>=0; i--)); do
    unset "unmapped_release_commits[${newly_mapped_indices[i]}]"
done
extra_commits=("${unmapped_release_commits[@]}") # What remains are the true extra commits

# 6. Determine missing commits
# First, remove all found original commits from the user's list
declare -a missing_commits=()
temp_user_commits=("${user_original_commits[@]}")
for mapped_orig in "${mapped_original_commits[@]}"; do
    for j in "${!temp_user_commits[@]}"; do
        if [ "${temp_user_commits[j]}" == "$mapped_orig" ]; then
            unset "temp_user_commits[j]"
            break
        fi
    done
done
missing_commits=("${temp_user_commits[@]}") # What remains is missing


# 7. Compare patches for all mapped cherry-picked commits.
for i in "${!mapped_release_commits[@]}"; do
    r_commit=${mapped_release_commits[i]}
    o_commit=${mapped_original_commits[i]}
    
    git format-patch --no-stat --stdout -1 "$o_commit" > "$TEMP_DIR/original.patch"
    git format-patch --no-stat --stdout -1 "$r_commit" > "$TEMP_DIR/release.patch"

    normalize_patch "$TEMP_DIR/original.patch" > "$TEMP_DIR/original.norm.patch"
    normalize_patch "$TEMP_DIR/release.patch" > "$TEMP_DIR/release.norm.patch"

    if ! diff_output=$(diff -u "$TEMP_DIR/original.norm.patch" "$TEMP_DIR/release.norm.patch"); then
        escaped_diff=$(echo "$diff_output" | json_escape)
        changed_commits_json_parts+=("{\"original_commit\": \"$o_commit\", \"release_commit\": \"$r_commit\", \"diff\": $escaped_diff}")
    fi
done

# 8. Assemble the final JSON output
status="VERIFICATION_SUCCESSFUL"
if [ ${#extra_commits[@]} -gt 0 ] || [ ${#missing_commits[@]} -gt 0 ] || [ ${#changed_commits_json_parts[@]} -gt 0 ]; then
    status="VERIFICATION_FAILED"
fi

extra_commits_json=$(printf '"%s",' "${extra_commits[@]}" | sed 's/,$//')
missing_commits_json=$(printf '"%s",' "${missing_commits[@]}" | sed 's/,$//')
changed_commits_json=$(IFS=,; echo "${changed_commits_json_parts[*]}")

cat <<EOF
{
    "status": "$status",
    "extra_commits": [$extra_commits_json],
    "missing_commits": [$missing_commits_json],
    "changed_commits": [$changed_commits_json]
}
EOF
