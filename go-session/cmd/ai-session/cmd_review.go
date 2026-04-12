package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/gemini"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/spf13/cobra"
)

var (
	reviewFlagRegular  bool
	reviewFlagDocs     bool
	reviewFlagDevOps   bool
	reviewFlagStrategy string
)

func init() {
	reviewCmd.Flags().BoolVar(&reviewFlagRegular, "regular", false, "Run regular code review (default when no flags given)")
	reviewCmd.Flags().BoolVar(&reviewFlagDocs, "docs", false, "Run documentation review")
	reviewCmd.Flags().BoolVar(&reviewFlagDevOps, "devops", false, "Run DevOps review")
	reviewCmd.Flags().StringVar(&reviewFlagStrategy, "strategy", "branch", "Diff strategy: branch (default) or last-commit")
	rootCmd.AddCommand(reviewCmd)
}

var reviewCmd = &cobra.Command{
	Use:   "review <story-id> [--regular] [--docs] [--devops] [--strategy=branch|last-commit]",
	Short: "Run headless LLM code review for a feature",
	Long: `Orchestrates LLM-based code review. Fetches the git diff in Go and
invokes the appropriate headless review prompt(s) via gemini --yolo.

Flag behaviour:
  No flags   → --regular is implied
  --docs     → docs review only (regular not included unless --regular also given)
  --devops   → devops review only (regular not included unless --regular also given)
  --regular  → explicit regular review (combinable with --docs and --devops)

Each type writes to its own file via 'ai-session review-write':
  regular → review.yml
  docs    → review-docs.yml
  devops  → review-devops.yml

Diff strategies (--strategy):
  branch      → (default) all changes in the current branch vs origin/<default-branch>,
                including committed, staged, unstaged, and untracked changes.
  last-commit → uncommitted changes only (staged, unstaged, untracked) vs HEAD.

Review types run sequentially.`,
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

		// No flags → imply --regular.
		anySpecial := reviewFlagDocs || reviewFlagDevOps
		runRegular := reviewFlagRegular || !anySpecial

		type reviewJob struct {
			enabled  bool
			typeName string
			prompt   string
		}
		jobs := []reviewJob{
			{runRegular, "regular", "review.md"},
			{reviewFlagDocs, "docs", "review-docs.md"},
			{reviewFlagDevOps, "devops", "review-devops.md"},
		}

		repoRoot := git.WorkDir()
		if repoRoot == "" {
			repoRoot = cwd
		}
		var diff string
		switch reviewFlagStrategy {
		case "branch":
			diff, err = git.Diff(repoRoot)
		case "last-commit":
			diff, err = git.DiffLastCommit(repoRoot)
		default:
			return fmt.Errorf("unknown strategy %q — must be one of: branch, last-commit", reviewFlagStrategy)
		}
		if err != nil {
			return fmt.Errorf("fetching git diff: %w", err)
		}

		aiHome := getAISessionHome()
		for _, job := range jobs {
			if !job.enabled {
				continue
			}
			fmt.Fprintf(os.Stderr, "Running %s review...\n", job.typeName)
			if err := runReviewJob(aiHome, featureDir, job.typeName, job.prompt, diff); err != nil {
				return fmt.Errorf("%s review failed: %w", job.typeName, err)
			}
		}
		return nil
	},
}

func runReviewJob(aiHome, featureDir, typeName, promptFile, diff string) error {
	promptPath := filepath.Join(aiHome, "headless", "session", promptFile)
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("reading prompt %s: %w", promptFile, err)
	}

	prompt := strings.ReplaceAll(string(promptBytes), "{{diff_here}}", diff)
	prompt = strings.ReplaceAll(prompt, "{{feature_dir_here}}", featureDir)
	prompt = strings.ReplaceAll(prompt, "{{review_type_here}}", typeName)

	return gemini.RunYolo(strings.NewReader(prompt), os.Stdout, os.Stderr)
}
