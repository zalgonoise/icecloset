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
	procID := fs.Int("pid", 0, "the process ID of the target process")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	if *procName == "" && *procID < 0 {
		return 1, ErrTargetProcessRequired
	}

	if *procID > 0 {
		if err := ResumeProcessByID(*procID); err != nil {
			return 1, err
		}

		return 0, nil
	}

	if err := ResumeProcessByName(ctx, *procName, logger); err != nil {
		return 1, err
	}

	return 0, nil
}

func ResumeProcessByName(ctx context.Context, procName string, logger *slog.Logger) error {
	h, err := proc.LoadProcessByName(ctx, procName, logger)
	if err != nil {
		return err
	}

	dll := windows.NewLazySystemDLL("ntdll.dll")

	if _, _, err := dll.NewProc("NtResumeProcess").Call(uintptr(h)); !isNilError(err) {
		return fmt.Errorf("failed to call NtResumeProcess: %w", err)
	}

	return nil
}

func ResumeProcessByID(procID int) error {
	h, err := proc.LoadProcessByID(procID)
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
