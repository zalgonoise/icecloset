package resume

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"

	"golang.org/x/sys/windows"

	"github.com/zalgonoise/icecloset/cmd/icecloset/proc"
)

var ErrTargetProcessRequired = errors.New("target process required")

func ExecResume(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("resume", flag.ExitOnError)

	procName := fs.String("name", "", "the name of the target process")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	if *procName == "" {
		return 1, ErrTargetProcessRequired
	}

	if err := Resume(ctx, *procName, logger); err != nil {
		return 1, err
	}

	return 0, nil
}

func Resume(ctx context.Context, procName string, logger *slog.Logger) error {
	h, err := proc.Load(ctx, procName, logger)
	if err != nil {
		return err
	}

	dll := windows.NewLazySystemDLL("ntdll.dll")

	if _, _, err := dll.NewProc("NtResumeProcess").Call(uintptr(h)); !isNilError(err) {
		return fmt.Errorf("failed to call NtResumeProcess: %w", err)
	}

	return nil
}

func isNilError(err error) bool {
	if err == nil {
		return true
	}

	if err.Error() == "The operation completed successfully." {
		return true
	}

	return false
}
