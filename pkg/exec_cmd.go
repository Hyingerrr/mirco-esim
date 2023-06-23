package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Hyingerrr/mirco-esim/log"
)

const (
	WireCmd = "wire"
)

type Exec interface {
	ExecWire(string, ...string) error

	ExecFmt(string, ...string) error

	ExecTest(string, ...string) error

	ExecBuild(string, ...string) error
}

type CmdExecOption func(*CmdExec)

type CmdExec struct {
	logger log.Logger
}

func NewCmdExec(options ...CmdExecOption) Exec {
	e := &CmdExec{}

	for _, option := range options {
		option(e)
	}

	if e.logger == nil {
		e.logger = log.NewNullLogger()
	}

	return e
}

func WithCmdExecLogger(logger log.Logger) CmdExecOption {
	return func(ce *CmdExec) {
		ce.logger = logger
	}
}

func (ce *CmdExec) ExecWire(dir string, args ...string) error {
	cmd := exec.Command(WireCmd, args...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	ce.logger.Infof("%s %s", WireCmd, strings.Join(args, ""))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}

func (ce *CmdExec) ExecFmt(dir string, args ...string) error {
	return nil
}

func (ce *CmdExec) ExecBuild(dir string, args ...string) error {
	cmdLine := fmt.Sprintf("build")

	args = append(strings.Split(cmdLine, " "), args...)

	ce.logger.Infof("go %s", strings.Join(args, ""))

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}

func (ce *CmdExec) ExecTest(dir string, args ...string) error {
	cmdLine := fmt.Sprintf("test -cover")

	args = append(strings.Split(cmdLine, " "), args...)

	ce.logger.Infof("go %s", strings.Join(args, " "))

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}

type NullExec struct{}

func NewNullExec() Exec {
	return &NullExec{}
}

func (ce *NullExec) ExecWire(dir string, args ...string) error {
	return nil
}

func (ce *NullExec) ExecFmt(dir string, args ...string) error {
	return nil
}

func (ce *NullExec) ExecBuild(dir string, args ...string) error {
	return nil
}

func (ce *NullExec) ExecTest(dir string, args ...string) error {
	return nil
}
