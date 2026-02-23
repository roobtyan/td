package app

import (
	"td/internal/config"
	"td/internal/timeutil"
)

type App struct {
	Config config.Config
	Clock  timeutil.Clock
}

func New(cfg config.Config, clk timeutil.Clock) *App {
	if clk == nil {
		clk = timeutil.SystemClock{}
	}
	return &App{
		Config: cfg,
		Clock:  clk,
	}
}
