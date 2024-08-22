package proc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/sys/windows"

	"github.com/zalgonoise/icecloset/perflib"
)

var ErrProcNotFound = errors.New("process not found")

type perfData struct {
	Name      string
	IDProcess float64 `perflib:"ID Process"`
}

func LoadProcessByName(ctx context.Context, procName string, logger *slog.Logger) (windows.Handle, error) {
	objs, err := perflib.GetPerflibSnapshot("")
	if err != nil {
		return 0, fmt.Errorf("failed to get perflib snapshot: %w", err)
	}

	data := make([]perfData, 0)
	if err := perflib.UnmarshalObject(ctx, objs["Process"], &data, logger); err != nil {
		return 0, fmt.Errorf("failed to fetch processes: %w", err)
	}

	for i := range data {
		if data[i].Name == procName {
			logger.InfoContext(ctx, "process", slog.String("process", data[i].Name), slog.Float64("proc_id", data[i].IDProcess), slog.Int("index", i))

			h, err := windows.OpenProcess(windows.PROCESS_SUSPEND_RESUME, false, uint32(data[i].IDProcess))
			if err != nil {
				return 0, fmt.Errorf("failed to open process: %w", err)
			}

			return h, nil
		}
	}

	return 0, ErrProcNotFound
}

func LoadProcessByID(procID int) (windows.Handle, error) {
	h, err := windows.OpenProcess(windows.PROCESS_SUSPEND_RESUME, false, uint32(procID))
	if err != nil {
		return 0, fmt.Errorf("failed to open process: %w", err)
	}

	return h, nil
}
