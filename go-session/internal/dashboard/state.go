package dashboard

// FeatureStatus is parsed directly from status.yaml.
type FeatureStatus struct {
	Mode         string `yaml:"mode"`
	Repo         string `yaml:"repo"`
	Branch       string `yaml:"branch"`
	PID          int    `yaml:"pid"`
	PipelineStep string `yaml:"pipeline_step"`
	StartedAt    string `yaml:"started_at"`
	UpdatedAt    string `yaml:"updated_at"`
}

// PlanTask is an unexported-friendly exported type for plan.yml task entries.
type PlanTask struct {
	ID     string `yaml:"id"`
	Status string `yaml:"status"`
}

// PlanSlice is an unexported-friendly exported type for plan.yml slice entries.
type PlanSlice struct {
	ID    string     `yaml:"id"`
	Tasks []PlanTask `yaml:"tasks"`
}

// FeatureState is the derived, template-ready view of one feature.
type FeatureState struct {
	StoryID      string
	Repo         string // org/repo — from status.yaml or derived from dir path
	Mode         string
	PipelineStep string
	IsRunning    bool
	LastDoneTask string // ID of last done task in document order
	AllDone      bool   // true if every task in plan.yml is "done"
	HasStatus    bool   // false if status.yaml was absent
}

// DeriveState computes a FeatureState from parsed inputs.
// isAlive is injected so callers can mock it in tests.
// status may be nil (feature has no status.yaml yet).
// repo is the fallback org/repo derived from the directory path.
func DeriveState(storyID string, repo string, status *FeatureStatus, plan []PlanSlice, isAlive func(int) bool) FeatureState {
	state := FeatureState{
		StoryID: storyID,
		Repo:    repo,
	}

	if status != nil {
		state.HasStatus = true
		state.Mode = status.Mode
		state.PipelineStep = status.PipelineStep
		state.IsRunning = status.PID > 0 && isAlive(status.PID)
		if status.Repo != "" {
			state.Repo = status.Repo
		}
	}

	if len(plan) == 0 {
		return state
	}

	var totalTasks int
	var lastDone string
	allDone := true
	for _, s := range plan {
		for _, t := range s.Tasks {
			totalTasks++
			if t.Status == "done" {
				lastDone = t.ID
			} else {
				allDone = false
			}
		}
	}
	state.AllDone = totalTasks > 0 && allDone
	state.LastDoneTask = lastDone

	return state
}
