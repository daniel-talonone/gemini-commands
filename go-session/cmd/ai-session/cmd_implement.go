package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/implement"
)

var (
	implementMaxRetries int
	implementRetryDelay time.Duration
)

func init() {
	implementCmd.Flags().IntVar(&implementMaxRetries, "max-retries", 5, "Maximum LLM+verification attempts per task")
	implementCmd.Flags().DurationVar(&implementRetryDelay, "retry-delay", 10*time.Second, "Delay between retry attempts (helps avoid rate limits)")
	implementCmd.Flags().Bool("tasks", false, "[deprecated] Execute one LLM call per task; use --strategy=task instead")
	implementCmd.Flags().Bool("slices", false, "[deprecated] Execute one LLM call per slice; use --strategy=slice instead")
	implementCmd.Flags().String("strategy", "", "Strategy to use: "+strings.Join(implement.KnownStrategies(), "|")+" (wins over --tasks/--slices)")
	rootCmd.AddCommand(implementCmd)
}

var implementCmd = &cobra.Command{
	Use:   "implement <story-id>",
	Short: "Start the headless implementation for a feature",
	Long: `Executes the headless implementation for a given feature.
This command replaces the 'orchestrate.sh --implement' script.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

		storyID := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current working directory: %w", err)
		}
		remoteURL := git.RemoteURL()

		featureDir, err := feature.ResolveFeatureDir(storyID, cwd, remoteURL)
		if err != nil {
			return fmt.Errorf("failed to resolve feature directory for %q: %w", storyID, err)
		}

		strategyFlag, _ := cmd.Flags().GetString("strategy")
		useSlices, _ := cmd.Flags().GetBool("slices")
		useTasks, _ := cmd.Flags().GetBool("tasks")

		var strategy implement.Strategy
		switch {
		case strategyFlag == "slice" || (strategyFlag == "" && useSlices):
			strategy = &implement.PerSliceStrategy{}
		case strategyFlag == "task" || strategyFlag == "" || useTasks:
			strategy = &implement.PerTaskStrategy{}
		default:
			return fmt.Errorf("unknown strategy %q: must be one of %s", strategyFlag, strings.Join(implement.KnownStrategies(), ", "))
		}

		runner, err := getRunner()
		if err != nil {
			return fmt.Errorf("invalid --model flag: %w", err)
		}

		if err := implement.Run(logger, storyID, featureDir, cwd, getAISessionHome(), implementMaxRetries, implementRetryDelay, strategy, runner); err != nil {
			return fmt.Errorf("implementation run failed: %w", err)
		}

		return nil
	},
}
