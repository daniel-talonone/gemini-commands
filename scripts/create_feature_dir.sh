#!/usr/bin/env bash
# Usage: ./create_feature_dir.sh <full-feature-path>
# Example: ./create_feature_dir.sh ~/.ai-session/features/org/repo/sc-12345

if [ -z "$1" ]; then
  echo "Error: Full feature directory path is required." >&2
  exit 1
fi

DIR_PATH="$1"

if [ -d "$DIR_PATH" ]; then
  echo "Warning: Directory $DIR_PATH already exists." >&2
else
  mkdir -p "$DIR_PATH"
  echo "Created directory: $DIR_PATH"
fi

# Create files with default content
echo "[]" > "$DIR_PATH/plan.yml"
echo "[]" > "$DIR_PATH/questions.yml"
echo "[]" > "$DIR_PATH/review.yml"

cat <<EOF > "$DIR_PATH/log.md"
# Work Log
*(This section is intentionally left blank.)*
EOF

cat <<EOF > "$DIR_PATH/pr.md"
# Pull Request
*(This section is intentionally left blank.)*
EOF

cat <<EOF > "$DIR_PATH/status.yaml"
mode: ''
repo: ''
branch: ''
pid: 0
pipeline_step: ''
started_at: ''
updated_at: ''
EOF

echo "Created and populated placeholder files in $DIR_PATH"
