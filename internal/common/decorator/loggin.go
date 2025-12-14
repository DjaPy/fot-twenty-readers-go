package decorator

import (
	"context"
	"log/slog"
)

type commandLoggingDecorator[C any] struct {
	base   CommandHandler[C]
	logger *slog.Logger
}

func (d commandLoggingDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	handlerType := generateActionName(cmd)

	logger := d.logger.With(slog.Group("command_fields",
		"command", handlerType,
		"command_body", cmd,
	))

	logger.Info("Executing command")
	defer func() {
		if err == nil {
			logger.Info("Command executed successfully")
		} else {
			logger.Error("Failed to execute command", "error", err)
		}
	}()

	return d.base.Handle(ctx, cmd) //nolint:wrapcheck // err repeated
}
