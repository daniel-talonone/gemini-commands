package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/daniel-talonone/gemini-commands/internal/description"
	"github.com/daniel-talonone/gemini-commands/internal/feature"
	"github.com/daniel-talonone/gemini-commands/internal/gemini"
	"github.com/daniel-talonone/gemini-commands/internal/git"
	"github.com/daniel-talonone/gemini-commands/internal/log"
	"github.com/daniel-talonone/gemini-commands/internal/plan"
	"github.com/daniel-talonone/gemini-commands/internal/pr"
	"github.com/daniel-talonone/gemini-commands/internal/status"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createPRDescriptionCmd)
}

var createPRDescriptionCmd = &cobra.Command{
	Use:   "create-pr-description <feature-name>",
	Short: "Generates a PR description and saves it to pr.md",
	Long: `Generates a PR description based on the feature's context and saves it to pr.md.

The command reads its inputs from:
- description.md (via description.LoadDescription)
- plan.yml (via plan package)
- log.md (via log package)
- work_dir and story_url from status.LoadStatus
- the PR template from <work_dir>/.github/pull_request_template.md (optional — skipped if absent)

The diff is fetched via git.Diff(workDir) (branch strategy).
All inputs are injected into headless/session/create-pr-description.md and piped to gemini --yolo,
which writes the result directly to pr.md via its file write tool.
If pr.md already has content, the command overwrites it (re-generation is idempotent).`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		featureName := args[0]
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current working directory: %w", err)
		}
		featureDir, err := feature.ResolveFeatureDir(featureName, cwd, git.RemoteURL())
		if err != nil {
			return err
		}

		s, err := status.LoadStatus(featureDir)
		if err != nil {
			return fmt.Errorf("loading status.yaml: %w", err)
		}

		repo := git.OrgRepo()
		branch := git.CurrentBranch()
		if err := status.Write(featureDir, "create-pr-description", repo, branch); err != nil {
			return fmt.Errorf("updating status: %w", err)
		}

		desc, err := description.LoadDescription(featureDir)
		if err != nil {
			return fmt.Errorf("loading description.md: %w", err)
		}

		p, err := plan.LoadPlan(featureDir)
		if err != nil {
			return fmt.Errorf("loading plan.yml: %w", err)
		}
		planStr, err := p.ToString()
		if err != nil {
			return fmt.Errorf("marshaling plan: %w", err)
		}

		l, err := log.LoadLog(featureDir)
		if err != nil {
			return fmt.Errorf("loading log.md: %w", err)
		}

		// Optional PR template — render as a section only when present.
		prTemplateSection := ""
		prTemplatePath := filepath.Join(s.WorkDir, ".github", "pull_request_template.md")
		if content, err := os.ReadFile(prTemplatePath); err == nil {
			prTemplateSection = "**Pull Request Template:**\n\n" + string(content) + "\n\n---\n"
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("reading PR template: %w", err)
		}

		// Optional story URL — render as a line only when present.
		storyURLSection := ""
		if s.StoryURL != "" {
			storyURLSection = "\n**Story URL:** " + s.StoryURL + "\n"
		}

		diff, err := git.Diff(s.WorkDir)
		if err != nil {
			return fmt.Errorf("getting git diff: %w", err)
		}

		promptTemplatePath := filepath.Join(getAISessionHome(), "headless", "session", "create-pr-description.md")
		promptTemplate, err := os.ReadFile(promptTemplatePath)
		if err != nil {
			return fmt.Errorf("reading prompt template: %w", err)
		}

		promptContent := strings.ReplaceAll(string(promptTemplate), "{{description_here}}", desc)
		promptContent = strings.ReplaceAll(promptContent, "{{plan_here}}", planStr)
		promptContent = strings.ReplaceAll(promptContent, "{{log_here}}", l)
		promptContent = strings.ReplaceAll(promptContent, "{{diff_here}}", diff)
		promptContent = strings.ReplaceAll(promptContent, "{{pr_template_section_here}}", prTemplateSection)
		promptContent = strings.ReplaceAll(promptContent, "{{story_url_section_here}}", storyURLSection)

		fmt.Fprintf(os.Stderr, "Generating PR description for feature %s...\n", featureName)
		var outputBuf bytes.Buffer
		if err := gemini.RunYolo(strings.NewReader(promptContent), io.MultiWriter(os.Stdout, &outputBuf), os.Stderr); err != nil {
			return fmt.Errorf("generating PR description: %w", err)
		}

		if err := pr.Write(featureDir, outputBuf.String()); err != nil {
			return fmt.Errorf("writing pr.md: %w", err)
		}

		if err := status.Write(featureDir, "pr-description-done", repo, branch); err != nil {
			return fmt.Errorf("updating status: %w", err)
		}
		return nil
	},
}
