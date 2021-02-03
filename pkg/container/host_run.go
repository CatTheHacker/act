package container

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/nektos/act/pkg/common"
)

// NewHostInput the input for the New function
type NewHostInput struct {
	WorkingDir string
}

// Host for managing tasks running on host
type Host interface {
	Create() common.Executor
	Copy(destPath string, files ...*FileEntry) common.Executor
	CopyDir(destPath string, srcPath string) common.Executor
	Exec(command []string, env map[string]string) common.Executor
	UpdateFromGithubEnv(env *map[string]string) common.Executor
	Remove() common.Executor
}

type hostReference struct {
	id    string
	input *NewHostInput
}

// NewHost creates a reference to a host
func NewHost(hostInput *NewHostInput) Host {
	hr := new(hostReference)
	hr.input = hostInput

	return hr
}

func (hr *hostReference) Create() common.Executor {
	return common.
		NewDebugExecutor("%susing host working dir=%s", logPrefix, hr.input.WorkingDir).
		Then(
			common.NewParallelExecutor(
				hr.executeCommand(fmt.Sprintf("mkdir -p %s", hr.input.WorkingDir), map[string]string{}),
				hr.executeCommand("pwd", map[string]string{}),
				hr.executeCommand("mkdir -p /actions/", map[string]string{}),
			))
}

func (hr *hostReference) executeCommand(command string, env map[string]string) common.Executor {
	return func(ctx context.Context) error {
		logger := common.Logger(ctx)

		words := strings.Fields(command)

		cmd := exec.Command(words[0], words[1:]...)
		cmd.Env = os.Environ()

		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		logger.Infof("  \u2601  running command: %s with env %s", command, env, cmd.Dir)

		return cmd.Run()
	}
}

func (hr *hostReference) Copy(destPath string, files ...*FileEntry) common.Executor {
	panic("implement me")
}

func (hr *hostReference) CopyDir(destPath string, srcPath string) common.Executor {
	return hr.executeCommand(fmt.Sprintf("cp -R %s %s", srcPath, destPath), map[string]string{})
}

func (hr *hostReference) Pull(forcePull bool) common.Executor {
	panic("implement me")
}

func (hr *hostReference) Start(attach bool) common.Executor {
	panic("implement me")
}

func (hr *hostReference) Exec(command []string, env map[string]string) common.Executor {
	envList := make([]string, 0)
	for k, v := range env {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}

	return hr.executeCommand(strings.Join(command, " "), env)
}

func (hr *hostReference) Remove() common.Executor {
	panic("implement me")
}

func (hr *hostReference) UpdateFromGithubEnv(env *map[string]string) common.Executor {
	return hr.extractGithubEnv(env).IfNot(common.Dryrun)
}

func (hr *hostReference) extractGithubEnv(env *map[string]string) common.Executor {
	if singleLineEnvPattern == nil {
		singleLineEnvPattern = regexp.MustCompile("^([^=]+)=([^=]+)$")
		mulitiLineEnvPattern = regexp.MustCompile(`^([^<]+)<<(\w+)$`)
	}

	localEnv := *env
	return func(ctx context.Context) error {
		l := common.Logger(ctx)
		f, err := os.Open(localEnv["GITHUB_ENV"])
		if err != nil {
			l.Errorf("Failed to read GITHUB_ENV: %v", err)
			return err
		}
		defer func() {
			err := f.Close()
			if err != nil {
				l.Errorf("Failed to close args file: %v", err)
			}
		}()
		s := bufio.NewScanner(f)
		multiLineEnvKey := ""
		multiLineEnvDelimiter := ""
		multiLineEnvContent := ""
		for s.Scan() {
			line := s.Text()
			if singleLineEnv := singleLineEnvPattern.FindStringSubmatch(line); singleLineEnv != nil {
				localEnv[singleLineEnv[1]] = singleLineEnv[2]
			}
			if line == multiLineEnvDelimiter {
				localEnv[multiLineEnvKey] = multiLineEnvContent
				multiLineEnvKey, multiLineEnvDelimiter, multiLineEnvContent = "", "", ""
			}
			if multiLineEnvKey != "" && multiLineEnvDelimiter != "" {
				if multiLineEnvContent != "" {
					multiLineEnvContent += "\n"
				}
				multiLineEnvContent += line
			}
			if mulitiLineEnvStart := mulitiLineEnvPattern.FindStringSubmatch(line); mulitiLineEnvStart != nil {
				multiLineEnvKey = mulitiLineEnvStart[1]
				multiLineEnvDelimiter = mulitiLineEnvStart[2]
			}
		}
		env = &localEnv
		return nil
	}
}
