package main

import (
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"net/http"

	"github.com/alecthomas/kong"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type cli struct {
	Loglevel string `help:"Set log level" default:"info" enum:"debug,info,warn,error"`
	Metrics  string `help:"Enable prometheus metrics on address" default:""`
	Run      RunCmd `cmd:"" default:"1" help:"Run the program"`
}

// configure slog logging
func setupLogging(level string) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	switch level {
	case "debug":
		opts.Level = slog.LevelDebug
	case "info":
		opts.Level = slog.LevelInfo
	case "warn":
		opts.Level = slog.LevelWarn
	case "error":
		opts.Level = slog.LevelError
	default:
		opts.Level = slog.LevelInfo
		slog.Error("invalid log level", "level", level)
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	buildInfo, _ := debug.ReadBuildInfo()
	child := logger.With(
		slog.Group("program_info",
			slog.Int("pid", os.Getpid()),
			slog.String("go_version", buildInfo.GoVersion),
		),
	)
	// log := slog.NewLogLogger(handler, slog.LevelError)
	slog.SetDefault(child)
}

func main() {
	var cli cli
	ctx := kong.Parse(&cli)
	setupLogging(cli.Loglevel)
	if cli.Metrics != "" {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			http.ListenAndServe(cli.Metrics, nil)
		}()
	}
	err := ctx.Run()
	if err != nil {
		slog.Error("error", "error", err)
	}
	ctx.FatalIfErrorf(err)
}

func waitfor() chan bool {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		<-sigs
		done <- true
	}()
	return done
}
