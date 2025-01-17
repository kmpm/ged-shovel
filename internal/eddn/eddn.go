package eddn

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/go-zeromq/zmq4"
)

const endpoint = "tcp://eddn.edcd.io:9500"

func Subscribe(ctx context.Context, wg *sync.WaitGroup, ch chan []byte) {
	defer wg.Done()

	sub := zmq4.NewSub(ctx)
	defer sub.Close()

	err := sub.Dial(endpoint)
	if err != nil {
		slog.Error("could not dial", "error", err)
		os.Exit(1)
	}

	err = sub.SetOption(zmq4.OptionSubscribe, "")
	if err != nil {
		slog.Error("could not subscribe", "error", err)
		os.Exit(1)
	}

	run := true
	wg.Add(1)
	go func() {
		<-ctx.Done()
		slog.Info("shutting down eddn subscriber")
		run = false
		wg.Done()
	}()

	for run {
		// Read envelope
		msg, err := sub.Recv()
		if err != nil {
			slog.Error("could not receive message", "error", err)
			// if eof then panic
			if errors.Is(err, io.EOF) {
				slog.Error("EOF received, exiting")
				panic(err)
			}
			slog.Error("other error, panic")
			panic(err)
		} else {
			switch msg.Type {
			case zmq4.UsrMsg:
				for _, frame := range msg.Frames {
					ch <- frame
				}
			case zmq4.CmdMsg:
				slog.Info("received command", "command", string(msg.Frames[0]), "framecount", len(msg.Frames))
			default:
				// slog.Info("received message", "type", msg.Type, "framecount", len(msg.Frames))
				slog.Error("unknown message type", "type", msg.Type, "framecount", len(msg.Frames))
			}
		}
	}
	slog.Info("eddn subscriber stopped")
}
