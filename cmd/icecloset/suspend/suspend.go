package suspend

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sys/windows"

	"github.com/zalgonoise/icecloset/cmd/icecloset/proc"
)

const defaultTimeout = 59 * time.Second

var ErrTargetProcessRequired = errors.New("target process required")

func ExecSuspend(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("resume", flag.ExitOnError)

	procName := fs.String("name", "", "the name of the target process")
	procID := fs.Int("pid", 0, "the process ID of the target process")
	timeout := fs.Duration("timeout", defaultTimeout, "the amount of time to wait before releasing the process")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	if *procName == "" && *procID < 0 {
		return 1, ErrTargetProcessRequired
	}

	if *timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	if *procID > 0 {
		if err := SuspendProcessByID(ctx, *procID); err != nil {
			return 1, err
		}

		return 0, nil
	}

	if err := SuspendProcessByName(ctx, *procName, logger); err != nil {
		return 1, err
	}

	return 0, nil
}

func SuspendProcessByName(ctx context.Context, procName string, logger *slog.Logger) error {
	h, err := proc.LoadProcessByName(ctx, procName, logger)
	if err != nil {
		return err
	}

	dll := windows.NewLazySystemDLL("ntdll.dll")

	if _, _, err := dll.NewProc("NtSuspendProcess").Call(uintptr(h)); !isNilError(err) {
		return fmt.Errorf("failed to call NtSuspendProcess: %w", err)
	}

	select {
	case <-ctx.Done():
		if _, _, err := dll.NewProc("NtResumeProcess").Call(uintptr(h)); !isNilError(err) {
			return fmt.Errorf("failed to call NtResumeProcess: %w", err)
		}
	}

	return nil
}

func SuspendProcessByID(ctx context.Context, procID int) error {
	h, err := proc.LoadProcessByID(procID)
	if err != nil {
		return err
	}

	dll := windows.NewLazySystemDLL("ntdll.dll")

	if _, _, err := dll.NewProc("NtSuspendProcess").Call(uintptr(h)); !isNilError(err) {
		return fmt.Errorf("failed to call NtSuspendProcess: %w", err)
	}

	select {
	case <-ctx.Done():
		if _, _, err := dll.NewProc("NtResumeProcess").Call(uintptr(h)); !isNilError(err) {
			return fmt.Errorf("failed to call NtResumeProcess: %w", err)
		}
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
