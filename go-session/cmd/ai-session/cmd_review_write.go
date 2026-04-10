package main

import (
	"fmt"
	"io"
	"os"

	"github.com/daniel-talonone/gemini-commands/internal/review"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var reviewWriteType string

func init() {
	reviewWriteCmd.Flags().StringVar(&reviewWriteType, "type", "", "Review type: regular, docs, devops (required)")
	_ = reviewWriteCmd.MarkFlagRequired("type")
	rootCmd.AddCommand(reviewWriteCmd)
}

var reviewWriteCmd = &cobra.Command{
	Use:   "review-write <feature-dir> --type <regular|docs|devops>",
	Short: "Validate and write review findings from stdin",
	Long: `Reads a YAML list of review findings from stdin, validates the schema,
and atomically replaces the appropriate review file in the feature directory.

This is the single write path for all review artifacts.
Prompts must never write review files directly.

Arguments:
  <feature-dir>  Path to the feature directory

Flags:
  --type  Review type: regular (→ review.yml), docs (→ review-docs.yml), devops (→ review-devops.yml)

Schema (each finding must have):
  id:       non-empty, kebab-case (e.g. "null-pointer-in-auth")
  file:     file path (may be empty string)
  line:     line number (may be 0)
  feedback: non-empty string describing the issue
  status:   "open" or "resolved"

On validation failure: exits non-zero; nothing is written to disk.
Error format: finding[N].field: <received> — <constraint> (e.g. <valid-example>)

Example:
  printf '%s' "$YAML" | ai-session review-write /path/to/feature --type regular`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		featureDir := args[0]

		t, err := parseReviewType(reviewWriteType)
		if err != nil {
			return err
		}

		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		var findings []review.Finding
		if err := yaml.Unmarshal(data, &findings); err != nil {
			return fmt.Errorf("invalid YAML input — %w\nEnsure the input is a valid YAML list of finding objects", err)
		}

		if err := review.Write(featureDir, t, findings); err != nil {
			return err
		}

		fmt.Printf("review file written successfully (%d findings).\n", len(findings))
		return nil
	},
}

// parseReviewType maps the --type flag value to a review.Type constant.
func parseReviewType(s string) (review.Type, error) {
	switch s {
	case "regular":
		return review.TypeDefault, nil
	case "docs":
		return review.TypeDocs, nil
	case "devops":
		return review.TypeDevOps, nil
	default:
		return "", fmt.Errorf("unknown review type %q — must be one of: regular, docs, devops", s)
	}
}
