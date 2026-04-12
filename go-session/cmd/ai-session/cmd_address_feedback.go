package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/gemini"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/github"
	"github.com/daniel-talonone/gemini-commands/internal/review"
	"github.com/spf13/cobra"
	"github.com/daniel-talonone/gemini-commands/internal/status"
)

var (
	addressFlagRegular bool
	addressFlagDocs    bool
	addressFlagDevOps  bool
	addressFlagRemote  bool
)

func init() {
	addressFeedbackCmd.Flags().BoolVar(&addressFlagRegular, "regular", false, "Address regular review findings")
	addressFeedbackCmd.Flags().BoolVar(&addressFlagDocs, "docs", false, "Address documentation review findings")
	addressFeedbackCmd.Flags().BoolVar(&addressFlagDevOps, "devops", false, "Address DevOps review findings")
	addressFeedbackCmd.Flags().BoolVar(&addressFlagRemote, "remote", false, "Address remote review findings from GitHub PR")
	rootCmd.AddCommand(addressFeedbackCmd)
}

var addressFeedbackCmd = &cobra.Command{
	Use:   "address-feedback <story-id> [--regular] [--docs] [--devops]",
	Short: "Address review findings for a feature using an LLM",
	Long: `Reads review findings from the feature directory and invokes gemini --yolo
to address them. The review package is the single source of truth for finding
locations and format — no file paths are hardcoded outside it.

Flag behaviour:
  No flags   → all review types are addressed (regular, docs, devops)
  --regular  → address regular review only
  --docs     → address documentation review only
  --devops   → address DevOps review only
  --remote   → address remote review findings from GitHub PR
  Flags are combinable: --regular --devops addresses both.
  --remote is mutually exclusive with other flags.

Types with no findings are skipped automatically.`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		storyID := args[0]
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting cwd: %w", err)
		}

		featureDir, err := feature.ResolveFeatureDir(storyID, cwd, git.RemoteURL())
		if err != nil {
			return fmt.Errorf("resolving feature dir: %w", err)
		}

		if addressFlagRemote {
			if addressFlagRegular || addressFlagDocs || addressFlagDevOps {
				return fmt.Errorf("--remote flag is mutually exclusive with --regular, --docs, and --devops")
			}

			s, err := status.LoadStatus(featureDir)
			if err != nil {
				return fmt.Errorf("loading status: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Fetching unresolved review threads from GitHub...\n")
			threads, err := github.GetUnresolvedReviewThreads(s.WorkDir, s.Branch)
			if err != nil {
				return fmt.Errorf("getting unresolved review threads: %w", err)
			}
			if threads == "" {
				fmt.Fprintf(os.Stderr, "no unresolved review threads found, skipping\n")
				if err := status.Write(featureDir, "feedback-remote-done", "", ""); err != nil {
					return fmt.Errorf("updating status: %w", err)
				}
				return nil
			}

			aiHome := getAISessionHome()
			promptPath := filepath.Join(aiHome, "headless", "session", "address-feedback-remote.md")
			promptBytes, err := os.ReadFile(promptPath)
			if err != nil {
				return fmt.Errorf("reading prompt: %w", err)
			}
			promptTemplate := string(promptBytes)

			prompt := strings.ReplaceAll(promptTemplate, "{{pr_comments_here}}", threads)
			prompt = strings.ReplaceAll(prompt, "{{feature_dir}}", featureDir)

			fmt.Fprintf(os.Stderr, "Addressing remote review findings...\n")
			if err := gemini.RunYolo(strings.NewReader(prompt), os.Stdout, os.Stderr); err != nil {
				return fmt.Errorf("gemini pipeline failed: %w", err)
			}

			if err := status.Write(featureDir, "feedback-remote-done", "", ""); err != nil {
				return fmt.Errorf("updating status: %w", err)
			}
			return nil
		}

		type addressJob struct {
			t        review.Type
			typeName string
		}

		var selectedTypes []review.Type
		anyFlag := addressFlagRegular || addressFlagDocs || addressFlagDevOps
		if anyFlag {
			if addressFlagRegular {
				selectedTypes = append(selectedTypes, review.TypeDefault)
			}
			if addressFlagDocs {
				selectedTypes = append(selectedTypes, review.TypeDocs)
			}
			if addressFlagDevOps {
				selectedTypes = append(selectedTypes, review.TypeDevOps)
			}
		} else {
			selectedTypes = review.AllTypes()
		}

		var jobs []addressJob
		for _, t := range selectedTypes {
			name, err := review.TypeName(t)
			if err != nil {
				return err
			}
			jobs = append(jobs, addressJob{t, name})
		}

		aiHome := getAISessionHome()
		promptPath := filepath.Join(aiHome, "headless", "session", "address-feedback.md")
		promptBytes, err := os.ReadFile(promptPath)
		if err != nil {
			return fmt.Errorf("reading prompt: %w", err)
		}
		promptTemplate := string(promptBytes)

		for _, job := range jobs {
			findings, err := review.ReadFindings(featureDir, job.t)
			if err != nil {
				return fmt.Errorf("reading %s findings: %w", job.typeName, err)
			}
			if findings == "" {
				fmt.Fprintf(os.Stderr, "no findings for %s review, skipping\n", job.typeName)
				continue
			}
			fmt.Fprintf(os.Stderr, "Addressing %s review findings...\n", job.typeName)
			if err := runAddressJob(promptTemplate, featureDir, job.typeName, findings); err != nil {
				return fmt.Errorf("addressing %s findings: %w", job.typeName, err)
			}
		}

		if err := status.Write(featureDir, "feedback-local-done", "", ""); err != nil {
			return fmt.Errorf("updating status: %w", err)
		}

		return nil
	},
}

func runAddressJob(promptTemplate, featureDir, typeName, findings string) error {
	prompt := strings.ReplaceAll(promptTemplate, "{{findings_here}}", findings)
	prompt = strings.ReplaceAll(prompt, "{{feature_dir}}", featureDir)
	prompt = strings.ReplaceAll(prompt, "{{review_type_here}}", typeName)

	return gemini.RunYolo(strings.NewReader(prompt), os.Stdout, os.Stderr)
}
