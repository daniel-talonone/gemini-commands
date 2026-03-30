#!/usr/bin/env bash
# Usage: ./create_feature_dir.sh <base-dir> <story-id>
# Example: ./create_feature_dir.sh .features sc-12345

if [ -z "$1" ] || [ -z "$2" ]; then
  echo "Error: Base directory and Story ID are required." >&2
  exit 1
fi

BASE_DIR="$1"
STORY_ID="$2"
DIR_PATH="$BASE_DIR/$STORY_ID"

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

echo "Created and populated placeholder files in $DIR_PATH"