package main

import (
	"github.com/zalgonoise/x/cli"

	"github.com/zalgonoise/icecloset/cmd/icecloset/resume"
	"github.com/zalgonoise/icecloset/cmd/icecloset/suspend"
)

var modes = []string{"suspend", "resume"}

func main() {
	runner := cli.NewRunner("icecloset",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"suspend": cli.Executable(suspend.ExecSuspend),
			"resume":  cli.Executable(resume.ExecResume),
		}),
	)

	cli.Run(runner)
}
