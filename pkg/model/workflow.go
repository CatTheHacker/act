package model

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nektos/act/pkg/common"
	"github.com/rhysd/actionlint"
	"github.com/wayneashleyberry/truecolor/pkg/color"

	log "github.com/sirupsen/logrus"
)

// ReadWorkflow returns a list of jobs for a given workflow file reader
func ReadWorkflow(in []byte) (Workflow, []*Error) {
	workflow, errs := actionlint.Parse(in)

	w := Workflow(*workflow)

	return w, errs
}

// CompositeRestrictions is the structure to control what is allowed in composite actions
type CompositeRestrictions struct {
	AllowCompositeUses            bool
	AllowCompositeIf              bool
	AllowCompositeContinueOnError bool
}

func defaultCompositeRestrictions() *CompositeRestrictions {
	return &CompositeRestrictions{
		AllowCompositeUses:            true,
		AllowCompositeIf:              true,
		AllowCompositeContinueOnError: false,
	}
}

type (
	Bool      = actionlint.Bool
	Error     = actionlint.Error
	Float     = actionlint.Float
	String    = actionlint.String
	Event     = actionlint.Event
	Container = actionlint.Container // ContainerSpec is the specification of the container to use for the job
	Strategy  = actionlint.Strategy  // Strategy for the job
	Defaults  = actionlint.Defaults  // Default settings that will apply to all steps in the job or workflow
	Output    = actionlint.Output
	Input     = actionlint.Input
	EnvVar    = actionlint.EnvVar

	Step            actionlint.Step     // Step is the structure of one step in a job
	Job             actionlint.Job      // Job is the structure of one job in a workflow
	Workflow        actionlint.Workflow // Workflow is the structure of the files in .github/workflows
	workflowPlanner struct {
		workflows map[string]*Workflow
	}
)

func (w *Workflow) String() string {
	if v := w.Name; v != nil {
		return v.Value
	}
	return ""
}

// GetJob will get a job by name in the workflow
func (w *Workflow) GetJob(jobID string) *Job {
	for id, j := range w.Jobs {
		if jobID == id {
			job := Job(*j)
			return &job
		}
	}
	return nil
}

// GetJobIDs will get all the job names in the workflow
func (w *Workflow) GetJobIDs() []string {
	ids := make([]string, 0)
	for id := range w.Jobs {
		ids = append(ids, id)
	}
	return ids
}

func (j *Job) GetNeeds() []string {
	var out []string
	for _, v := range j.Needs {
		if v != nil {
			out = append(out, v.Value)
		} else {
			log.Debug("found nil value in job.Needs")
		}
	}
	return out
}

func (j *Job) GetRunsOn() []string {
	var out []string
	if r := j.RunsOn; r != nil {
		for _, v := range j.RunsOn.Labels {
			if v != nil {
				out = append(out, v.Value)
			} else {
				log.Debug("found nil value in job.RunsOn.Labels")
			}
		}
	}
	return out
}

func (j *Job) GetContainer() *Container {
	if v := j.Container; v != nil {
		return v
	}
	return nil
}

func (j *Job) GetMatrixes() []map[string]string {
	var outcome []map[string]string
	if j.Strategy != nil {
		if j.Strategy.Matrix != nil {
			var matrixes, excludes, includes []map[string]actionlint.RawYAMLValue

			if j.Strategy.Matrix.Rows != nil {
				log.Debugf("%s:", color.Color(255, 0, 89).Sprint("j.Strategy.Matrix.Rows"))
				common.LogMatrixRows(j.Strategy.Matrix.Rows)

				matrixes = common.CartesianProduct(j.Strategy.Matrix.Rows)
			}

			if j.Strategy.Matrix.Exclude != nil && j.Strategy.Matrix.Exclude.Combinations != nil {
				log.Debug()
				log.Debugf("%s:", color.Color(255, 0, 89).Sprint("j.Strategy.Matrix.Exclude.Combinations"))

				excludeCombinations := ConvertCombinations(j.Strategy.Matrix.Exclude.Combinations)

				common.LogMatrixRows(excludeCombinations)

				excludes = common.CartesianProduct(excludeCombinations)
			}

			if j.Strategy.Matrix.Include != nil && j.Strategy.Matrix.Include.Combinations != nil {
				log.Debug()
				log.Debugf("%s:", color.Color(255, 0, 89).Sprint("j.Strategy.Matrix.Include.Combinations"))

				includeCombinations := ConvertCombinations(j.Strategy.Matrix.Include.Combinations)

				common.LogMatrixRows(includeCombinations)

				includes = common.CartesianProduct(includeCombinations)
			}

		MATRIX:
			for _, matrix := range matrixes {
				for _, exclude := range excludes {
					// if commonKeysMatch(matrix, exclude) {
					// 	log.Debugf("Skipping matrix '%v' due to exclude '%v'", matrix, exclude)
					// 	continue MATRIX
					// }
					if commonKeysMatchString(ConvertCartesianProduct(matrix), ConvertCartesianProduct(exclude)) {
						log.Debugf("Skipping matrix '%v' due to exclude '%v'",
							color.Color(210, 36, 41).Sprint(matrix),
							color.Color(74, 155, 155).Sprint(exclude),
						)
						continue MATRIX
					}
				}
				outcome = append(outcome, ConvertCartesianProduct(matrix))
			}
			for _, include := range includes {
				log.Debugf("Adding include '%v'", include)
				outcome = append(outcome, ConvertCartesianProduct(include))
			}
		}
	}
	return outcome
}

func ConvertCombinations(c []*actionlint.MatrixCombination) map[string]*actionlint.MatrixRow {
	out := make(map[string]*actionlint.MatrixRow)
	for _, v := range c {
		for k, v := range v.Assigns {
			out[k] = &actionlint.MatrixRow{Name: &actionlint.String{Value: k}, Values: []actionlint.RawYAMLValue{v.Value}}
		}
	}
	return out
}

func ConvertCartesianProduct(c map[string]actionlint.RawYAMLValue) map[string]string {
	out := make(map[string]string)
	for k, v := range c {
		out[k] = strings.Trim(v.String(), `"`)
	}
	return out
}

func commonKeysMatchString(a map[string]string, b map[string]string) bool {
	log.Debugf("%s:", color.Color(255, 0, 89).Sprint("Common keys match"))
	for aKey, aVal := range a {
		log.Debugf("\t%s:", color.Color(255, 165, 00).Sprint(aKey))
		log.Debugf("\t\t%s",
			color.Color(255, 165, 255).Sprint(aVal),
		)
		if bVal, ok := b[aKey]; ok && !reflect.DeepEqual(aVal, bVal) {
			log.Debugf("\t\t%s",
				color.Color(255, 165, 255).Sprint(bVal),
			)
			return false
		}
	}
	return true
}

// nolint:unused,deadcode
func commonKeysMatch(a map[string]actionlint.RawYAMLValue, b map[string]actionlint.RawYAMLValue) bool {
	log.Debugf("%s:", color.Color(255, 0, 89).Sprint("Common keys match"))
	for aKey, aVal := range a {
		log.Debugf("\t%s:", color.Color(255, 165, 00).Sprint(aKey))
		log.Debugf("\t\t%s",
			color.Color(255, 165, 255).Sprint(aVal),
		)
		if bVal, ok := b[aKey]; ok && !reflect.DeepEqual(aVal, bVal) {
			log.Debugf("\t\t%s",
				color.Color(255, 165, 255).Sprint(bVal),
			)
			return false
		}
	}
	return true
}

func (s *Step) GetName() string {
	if v := s.Name; v != nil {
		return v.Value
	}
	return ""
}

func (s *Step) GetID() string {
	if v := s.ID; v != nil {
		return v.Value
	}
	return ""
}

// String gets the name of step
func (s *Step) String() string {
	if v := s.GetName(); v != "" {
		return v
	} else if v := s.Uses(); v != "" {
		return v
	} else if v := s.Run(); v != "" {
		return v
	}
	return s.GetID()
}

func (s *Step) With() map[string]string {
	out := make(map[string]string)
	if s.StepTypeCheck() == ExecKindAction {
		for k, v := range s.ExecAction().Inputs {
			if v != nil {
				out[k] = v.Value.Value
			} else {
				log.Debugf("found %s = nil value in step.With", k)
			}
		}
	}
	return out
}

func (s *Step) GetTimeout() float64 {
	if v := s.TimeoutMinutes; v != nil {
		return v.Value
	}
	return 0
}

func (s *Step) GetContinueOnError() bool {
	if v := s.ContinueOnError; v != nil {
		return v.Value
	}
	return false
}

const (
	ExecKindAction = iota
	ExecKindRun
)

type ExecKind uint8

func (s *Step) StepTypeCheck() ExecKind {
	if s != nil {
		step := actionlint.Step(*s)
		switch step.Exec.Kind() {
		case actionlint.ExecKindAction:
			return ExecKindAction
		case actionlint.ExecKindRun:
			return ExecKindRun
		}
	}
	return 255
}

type ExecAction actionlint.ExecAction

func (s *Step) ExecAction() *actionlint.ExecAction {
	return s.Exec.(*actionlint.ExecAction)
}

func (s *Step) Uses() string {
	if s.StepTypeCheck() == ExecKindAction {
		return s.ExecAction().Uses.Value
	}
	return ""
}

type ExecRun actionlint.ExecRun

func (s *Step) ExecRun() *actionlint.ExecRun {
	return s.Exec.(*actionlint.ExecRun)
}

func (s *Step) Run() string {
	if r := s.ExecRun().Run; r != nil {
		return r.Value
	}
	return ""
}

func (s *Step) Shell() string {
	if s := s.ExecRun().Shell; s != nil {
		return s.Value
	}
	return ""
}

func (s *Step) WorkingDirectory() string {
	if wd := s.ExecRun().WorkingDirectory; wd != nil {
		return wd.Value
	}
	return ""
}

// ShellCommand returns the command for the shell
func (s *Step) ShellCommand() string {
	//Reference: https://github.com/actions/runner/blob/8109c962f09d9acc473d92c595ff43afceddb347/src/Runner.Worker/Handlers/ScriptHandlerHelpers.cs#L9-L17
	switch s.Shell() {
	case "", "bash":
		return "bash --noprofile --norc -e -o pipefail {0}"
	case "pwsh":
		return "pwsh -command . '{0}'"
	case "python":
		return "python {0}"
	case "sh":
		return "sh -e -c {0}"
	case "cmd":
		return "%ComSpec% /D /E:ON /V:OFF /S /C \"CALL \"{0}\"\""
	case "powershell":
		return "powershell -command . '{0}'"
	default:
		return s.Shell()
	}
}

type StepType int // StepType describes what type of step we are about to run

const (
	// StepTypeRun is all steps that have a `run` attribute
	StepTypeRun StepType = iota

	// StepTypeUsesDockerURL is all steps that have a `uses` that is of the form `docker://...`
	StepTypeUsesDockerURL

	// StepTypeUsesActionLocal is all steps that have a `uses` that is a local action in a subdirectory
	StepTypeUsesActionLocal

	// StepTypeUsesActionRemote is all steps that have a `uses` that is a reference to a github repo
	StepTypeUsesActionRemote

	// StepTypeInvalid is for steps that have invalid step action
	StepTypeInvalid
)

func (s *Step) Type() StepType {
	if s.StepTypeCheck() == ExecKindRun {
		return StepTypeRun
	}
	if s.StepTypeCheck() == ExecKindAction {
		if strings.HasPrefix(s.Uses(), "docker://") {
			return StepTypeUsesDockerURL
		}
		if strings.HasPrefix(s.Uses(), "./") {
			return StepTypeUsesActionLocal
		}
		return StepTypeUsesActionRemote
	}
	return StepTypeInvalid
}

func (s *Step) Validate(config *CompositeRestrictions) error {
	if config == nil {
		config = defaultCompositeRestrictions()
	}
	if s.Type() != StepTypeRun && !config.AllowCompositeUses {
		return fmt.Errorf("(StepID: %s): Unexpected value 'uses'", s.String())
	} else if s.Type() == StepTypeRun && s.Shell() == "" {
		return fmt.Errorf("(StepID: %s): Required property is missing: 'shell'", s.String())
	} else if s.If != nil && !config.AllowCompositeIf {
		return fmt.Errorf("(StepID: %s): Property is not available: 'if'", s.String())
	} else if s.ContinueOnError.Value && !config.AllowCompositeContinueOnError {
		return fmt.Errorf("(StepID: %s): Property is not available: 'continue-on-error'", s.String())
	}
	return nil
}

func ConvertMap(env map[string]*EnvVar) map[string]string {
	out := make(map[string]string)
	for _, v := range env {
		log.Debug(v)
		out[v.Name.Value] = v.Value.Value
	}
	return out
}
