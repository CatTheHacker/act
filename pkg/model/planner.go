package model

import (
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"

	log "github.com/sirupsen/logrus"
)

// WorkflowPlanner contains methods for creating plans
type WorkflowPlanner interface {
	PlanEvent(eventName string) *Plan
	PlanJob(jobName string) *Plan
	GetEvents() []Event
}

// Plan contains a list of stages to run in series
type Plan struct {
	Stages []*Stage
}

// Stage contains a list of runs to execute in parallel
type Stage struct {
	Runs []*Run
}

// Run represents a job from a workflow that needs to be run
type Run struct {
	File     string
	JobID    string
	Workflow *Workflow
}

func (r *Run) String() string {
	if r.Job().Name == nil {
		return r.JobID
	}
	return r.Job().Name.Value
}

// Job returns the job for this Run
func (r *Run) Job() *Job {
	return r.Workflow.GetJob(r.JobID)
}

type WorkflowFiles struct {
	workflowFileInfo os.FileInfo
	dirPath          string
}

// NewWorkflowPlanner will load a specific workflow, all workflows from a directory or all workflows from a directory and its subdirectories
// nolint: gocyclo
func NewWorkflowPlanner(path string) (WorkflowPlanner, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var workflows []WorkflowFiles

	if fi.IsDir() {
		log.Debugf("Loading workflows from '%s'", path)
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, v := range files {
			workflows = append(workflows, WorkflowFiles{
				dirPath:          path,
				workflowFileInfo: v,
			})
		}
	} else {
		log.Debugf("Loading workflow '%s'", path)
		dirname := filepath.Dir(path)

		workflows = append(workflows, WorkflowFiles{
			dirPath:          dirname,
			workflowFileInfo: fi,
		})
	}
	if err != nil {
		return nil, err
	}

	wp := new(workflowPlanner)
	wp.workflows = map[string]*Workflow{}
	for _, wf := range workflows {
		ext := filepath.Ext(wf.workflowFileInfo.Name())
		if ext == ".yml" || ext == ".yaml" {
			f, err := os.ReadFile(filepath.Join(wf.dirPath, wf.workflowFileInfo.Name()))
			if err != nil {
				return nil, err
			}

			log.Debugf("Reading workflow '%s'", wf.workflowFileInfo.Name())
			w, errs := ReadWorkflow(f)
			log.Errorf("%#v", errs)

			wp.workflows[wf.workflowFileInfo.Name()] = &w
		}
	}

	return wp, nil
}

// PlanEvent builds a new list of runs to execute in parallel for an event name
func (wp *workflowPlanner) PlanEvent(eventName string) *Plan {
	plan := new(Plan)
	if len(wp.workflows) == 0 {
		log.Debugf("no events found for workflow: %s", eventName)
	}

	for f, w := range wp.workflows {
		for _, e := range w.On {
			if e.EventName() == eventName {
				plan.mergeStages(createStages(f, w, w.GetJobIDs()...))
			}
		}
	}
	return plan
}

// PlanJob builds a new run to execute in parallel for a job name
func (wp *workflowPlanner) PlanJob(jobName string) *Plan {
	plan := new(Plan)
	if len(wp.workflows) == 0 {
		log.Debugf("no jobs found for workflow: %s", jobName)
	}

	for f, w := range wp.workflows {
		plan.mergeStages(createStages(f, w, jobName))
	}
	return plan
}

// GetEvents gets all the events in the workflows file
func (wp *workflowPlanner) GetEvents() []Event {
	events := make([]Event, 0)
	for _, w := range wp.workflows {
		found := false
		for _, e := range events {
			for _, we := range w.On {
				if e == we {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			events = append(events, w.On...)
		}
	}

	// sort the list based on depth of dependencies
	sort.Slice(events, func(i, j int) bool {
		return events[i].EventName() < events[j].EventName()
	})

	return events
}

// MaxRunNameLen determines the max name length of all jobs
func (p *Plan) MaxRunNameLen() int {
	maxRunNameLen := 0
	for _, stage := range p.Stages {
		for _, run := range stage.Runs {
			runNameLen := len(run.String())
			if runNameLen > maxRunNameLen {
				maxRunNameLen = runNameLen
			}
		}
	}
	return maxRunNameLen
}

// GetJobIDs will get all the job names in the stage
func (s *Stage) GetJobIDs() []string {
	names := make([]string, 0)
	for _, r := range s.Runs {
		names = append(names, r.JobID)
	}
	return names
}

// Merge stages with existing stages in plan
func (p *Plan) mergeStages(stages []*Stage) {
	newStages := make([]*Stage, int(math.Max(float64(len(p.Stages)), float64(len(stages)))))
	for i := 0; i < len(newStages); i++ {
		newStages[i] = new(Stage)
		if i >= len(p.Stages) {
			newStages[i].Runs = append(newStages[i].Runs, stages[i].Runs...)
		} else if i >= len(stages) {
			newStages[i].Runs = append(newStages[i].Runs, p.Stages[i].Runs...)
		} else {
			newStages[i].Runs = append(newStages[i].Runs, p.Stages[i].Runs...)
			newStages[i].Runs = append(newStages[i].Runs, stages[i].Runs...)
		}
	}
	p.Stages = newStages
}

func createStages(f string, w *Workflow, jobIDs ...string) []*Stage {
	// first, build a list of all the necessary jobs to run, and their dependencies
	jobDependencies := make(map[string][]string)
	for len(jobIDs) > 0 {
		newJobIDs := make([]string, 0)
		for _, jID := range jobIDs {
			// make sure we haven't visited this job yet
			if _, ok := jobDependencies[jID]; !ok {
				if job := w.GetJob(jID); job != nil {
					jobDependencies[jID] = job.GetNeeds()
					newJobIDs = append(newJobIDs, job.GetNeeds()...)
				}
			}
		}
		jobIDs = newJobIDs
	}

	// next, build an execution graph
	stages := make([]*Stage, 0)
	for len(jobDependencies) > 0 {
		stage := new(Stage)
		for jID, jDeps := range jobDependencies {
			// make sure all deps are in the graph already
			if listInStages(jDeps, stages...) {
				stage.Runs = append(stage.Runs, &Run{
					File:     f,
					JobID:    jID,
					Workflow: w,
				})
				delete(jobDependencies, jID)
			}
		}
		if len(stage.Runs) == 0 {
			log.Fatalf("Unable to build dependency graph!")
		}
		stages = append(stages, stage)
	}

	return stages
}

// return true iff all strings in srcList exist in at least one of the stages
func listInStages(srcList []string, stages ...*Stage) bool {
	for _, src := range srcList {
		found := false
		for _, stage := range stages {
			for _, search := range stage.GetJobIDs() {
				if src == search {
					found = true
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}
