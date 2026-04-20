#!/usr/bin/env bash
# Smoke test for scripts/orchestrate.sh.
# Tests arg parsing, --status, and precondition checks.
# Does NOT invoke gemini — stops at the precondition boundary.
#
# Run from anywhere:
#   scripts/test_orchestrate.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ORCHESTRATE="$SCRIPT_DIR/orchestrate.sh"
FEATURE_ID="smoke-test-orchestrate-$$"
FEATURE_DIR="$SCRIPT_DIR/../.features/$FEATURE_ID"

cd "$SCRIPT_DIR/.."
mkdir -p "$FEATURE_DIR"
trap 'rm -rf "$FEATURE_DIR"' EXIT

pass=0
fail=0

check() {
  local desc="$1"
  local expected_exit="$2"
  shift 2
  local actual_exit=0
  "$@" &>/dev/null || actual_exit=$?
  if [ "$actual_exit" -eq "$expected_exit" ]; then
    echo "  PASS  $desc"
    pass=$((pass + 1))
  else
    echo "  FAIL  $desc (expected exit $expected_exit, got $actual_exit)"
    fail=$((fail + 1))
  fi
}

check_stderr() {
  local desc="$1"
  local pattern="$2"
  shift 2
  local stderr_out
  stderr_out="$("$@" 2>&1 1>/dev/null || true)"
  if echo "$stderr_out" | grep -q "$pattern"; then
    echo "  PASS  $desc"
    pass=$((pass + 1))
  else
    echo "  FAIL  $desc (pattern '$pattern' not in stderr: '$stderr_out')"
    fail=$((fail + 1))
  fi
}

echo "==> smoke test: orchestrate.sh"
echo ""

# --- Arg parsing ---
echo "-- arg parsing --"
check "no args → exit 1"                        1  "$ORCHESTRATE"
check "missing flag → exit 1"                   1  "$ORCHESTRATE" "$FEATURE_ID"
check "unknown flag → exit 1"                   1  "$ORCHESTRATE" "$FEATURE_ID" "--bad"
check "--status missing story-id → exit 1"      1  "$ORCHESTRATE" "--status"

# --- --status ---
echo ""
echo "-- --status --"
check "--status nonexistent dir → exit 1"       1  "$ORCHESTRATE" "--status" "no-such-story-xyz"
check "--status no status.yaml → exit 1"        1  "$ORCHESTRATE" "--status" "$FEATURE_ID"

printf 'mode: auto\nrepo: org/repo\nbranch: %s\npid: 1\npipeline_step: plan\nstarted_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n' \
  "$FEATURE_ID" > "$FEATURE_DIR/status.yaml"
check        "--status valid → exit 0"          0  "$ORCHESTRATE" "--status" "$FEATURE_ID"
# check stdout of --status contains pipeline_step
status_out="$("$ORCHESTRATE" "--status" "$FEATURE_ID" 2>/dev/null)"
if echo "$status_out" | grep -q "pipeline_step"; then
  echo "  PASS  --status outputs pipeline_step"
  pass=$((pass + 1))
else
  echo "  FAIL  --status outputs pipeline_step (got: $status_out)"
  fail=$((fail + 1))
fi

# --- Preconditions: no description.md ---
echo ""
echo "-- preconditions: no description.md --"
for flag in --plan --implement --review --pr; do
  check_stderr "$flag no description.md → mentions description.md" \
    "description.md" "$ORCHESTRATE" "$FEATURE_ID" "$flag"
done

# --- Preconditions: description.md present, no plan.yml ---
echo ""
echo "-- preconditions: description.md present --"
printf '# smoke test feature\n' > "$FEATURE_DIR/description.md"

for flag in --implement --review --pr; do
  check_stderr "$flag no plan.yml → mentions plan.yml" \
    "plan.yml" "$ORCHESTRATE" "$FEATURE_ID" "$flag"
done

# --- Preconditions: plan.yml present ---
echo ""
echo "-- preconditions: plan.yml present --"
printf '%s\n' '- id: s1' '  description: test' '  status: done' '  tasks: []' > "$FEATURE_DIR/plan.yml"

check_stderr "--review no IMPLEMENT COMPLETE in log → mentions implement" \
  "implement" "$ORCHESTRATE" "$FEATURE_ID" "--review"

check_stderr "--pr no review.yml → mentions review.yml" \
  "review.yml" "$ORCHESTRATE" "$FEATURE_ID" "--pr"

# --- Preconditions satisfied: should reach gemini (not fail on preconditions) ---
echo ""
echo "-- preconditions fully satisfied --"
printf '## [2026-01-01T00:00:00Z]\nIMPLEMENT COMPLETE: All slices done.\n' > "$FEATURE_DIR/log.md"
printf '%s\n' '- id: f1' '  file: x' '  feedback: y' '  status: open' > "$FEATURE_DIR/review.yml"

for flag in --review --pr; do
  # Should NOT fail with a precondition error — error (if any) comes from gemini.
  # Timeout after 5s to avoid waiting on gemini startup.
  local_exit=0
  stderr_out="$(timeout 5 "$ORCHESTRATE" "$FEATURE_ID" "$flag" 2>&1 1>/dev/null || local_exit=$?)"
  if echo "$stderr_out" | grep -qE "description\.md|plan\.yml|review\.yml|implement"; then
    echo "  FAIL  $flag preconditions satisfied but got precondition error: $stderr_out"
    fail=$((fail + 1))
  else
    echo "  PASS  $flag preconditions satisfied (reached gemini)"
    pass=$((pass + 1))
  fi
done

# --- Summary ---
echo ""
echo "==> $pass passed, $fail failed"
[ "$fail" -eq 0 ] && exit 0 || exit 1
