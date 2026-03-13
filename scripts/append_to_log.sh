#!/bin/bash

# Appends a timestamped message to a log file.
#
# Usage:
#   ./append_to_log.sh <log_file_path> <log_message>
#
# Arguments:
#   $1 (log_file_path): The path to the log file (e.g., .vscode/sc-12345/log.md).
#   $2 (log_message): The message to append.

set -e

LOG_FILE="$1"
LOG_MESSAGE="$2"

if [ -z "$LOG_FILE" ] || [ -z "$LOG_MESSAGE" ]; then
  echo "Error: Both log file path and log message are required." >&2
  echo "Usage: $0 <log_file_path> <log_message>" >&2
  exit 1
fi

if [ ! -f "$LOG_FILE" ]; then
  # If the file doesn't exist, create it.
  touch "$LOG_FILE"
fi

# Generate a timestamp in ISO 8601 format (UTC)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Format the log entry
# Use printf for robust handling of special characters in LOG_MESSAGE
# Add a newline before the header if the file is not empty
if [ -s "$LOG_FILE" ]; then
    printf "
" >> "$LOG_FILE"
fi
printf "## [%s]

%s
" "$TIMESTAMP" "$LOG_MESSAGE" >> "$LOG_FILE"

echo "Log entry added to $LOG_FILE"
