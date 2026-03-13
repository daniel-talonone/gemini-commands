#!/bin/bash

# A robust script to migrate a single-file feature document to the new
# multi-file directory structure.

set -e # Exit immediately if a command exits with a non-zero status.
set -u # Treat unset variables as an error

# --- Argument Validation ---
if [ -z "$1" ]; then
  echo "Error: No feature document path provided." >&2
  echo "Usage: $0 /path/to/feature.md" >&2
  exit 1
fi

OLD_FILE_PATH=$1

if [ ! -f "$OLD_FILE_PATH" ]; then
  echo "Error: File not found at '$OLD_FILE_PATH'" >&2
  exit 1
fi

# --- Path Derivation ---
FEATURE_DIR=$(echo "$OLD_FILE_PATH" | sed 's/\.md$//')
ARCHIVE_PATH="${OLD_FILE_PATH}.migrated"
BASENAME=$(basename "$FEATURE_DIR")

echo "Starting migration for $BASENAME..."
echo "Old file: $OLD_FILE_PATH"
echo "New directory: $FEATURE_DIR"

# --- Directory and File Creation ---
mkdir -p "$FEATURE_DIR"
echo "Created directory: $FEATURE_DIR"

# Create all potential files to ensure they exist even if sections are empty
touch "$FEATURE_DIR/description.md"
touch "$FEATURE_DIR/plan.yml"
touch "$FEATURE_DIR/questions.yml"
touch "$FEATURE_DIR/review.yml"
touch "$FEATURE_DIR/log.md"
touch "$FEATURE_DIR/pr.md"

# --- Content Parsing and File Population ---

# Use awk to parse sections. This is more efficient and robust than multiple passes.
awk -v dir="$FEATURE_DIR" '
  function generate_id(text) {
    # Simple kebab-case ID generator from the first 50 chars
    id = tolower(text)
    gsub(/[^a-z0-9 ]/, "", id)
    gsub(/[ ]+/, "-", id)
    id = substr(id, 0, 50)
    gsub(/-$/, "", id); # Remove trailing hyphen
    return id
  }

  # Default state: everything goes to description.md
  BEGIN { current_section = "description" }

  # Switch section based on headers
  /^## Next Steps/ { current_section = "plan"; next }
  /^## Open Questions/ { current_section = "questions"; next }
  /^## Code Review Feedback/ { current_section = "review"; next }
  /^## (Work Log|Checkpoints)/ { current_section = "log"; next }
  /^## Pull Request/ { current_section = "pr"; next }
  
  # Ignore content under any other h2 heading
  /^## / { current_section = "ignore"; next }

  # Process content based on the current section
  {
    # Skip empty lines
    if ($0 ~ /^[[:space:]]*$/) next;

    if (current_section == "description") {
      print >> (dir "/description.md")
    } else if (current_section == "log") {
      print >> (dir "/log.md")
    } else if (current_section == "pr") {
      print >> (dir "/pr.md")
    } else if (current_section == "plan" && /^- /) {
      task = substr($0, 3)
      gsub(/"/, """, task) # Escape quotes
      id = generate_id(task)
      print "- id: " id >> (dir "/plan.yml")
      print "  task: "" task """ >> (dir "/plan.yml")
      print "  status: todo" >> (dir "/plan.yml")
    } else if (current_section == "questions" && /^- /) {
      question = substr($0, 3)
      gsub(/"/, """, question) # Escape quotes
      id = generate_id(question)
      print "- id: " id >> (dir "/questions.yml")
      print "  question: "" question """ >> (dir "/questions.yml")
      print "  status: open" >> (dir "/questions.yml")
      print "  answer: null" >> (dir "/questions.yml")
    } else if (current_section == "review" && /^- /) {
      line_content = substr($0, 3)
      # Match file:line:feedback format, robustly handling colons in feedback
      if (match(line_content, /^([^:]+):([0-9]+):(.*)$/, parts)) {
        file = parts[1]
        line = parts[2]
        feedback = parts[3]
        
        # Trim whitespace
        gsub(/^[ 	]+|[ 	]+$/, "", file);
        gsub(/^[ 	]+|[ 	]+$/, "", line);
        gsub(/^[ 	]+|[ 	]+$/, "", feedback);
        gsub(/"/, """, feedback)

        id = generate_id(feedback)
        print "- id: " id >> (dir "/review.yml")
        print "  file: " file >> (dir "/review.yml")
        print "  line: " line >> (dir "/review.yml")
        print "  feedback: "" feedback """ >> (dir "/review.yml")
        print "  status: open" >> (dir "/review.yml")
      }
    }
  }
' "$OLD_FILE_PATH"

echo "Parsed content and populated new files."

# --- Archive Old File ---
mv "$OLD_FILE_PATH" "$ARCHIVE_PATH"
echo "Archived old file to $ARCHIVE_PATH"

# --- Final Confirmation ---
echo ""
echo "✅ Migration for $BASENAME complete!"
