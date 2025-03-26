package common

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// Struct that simulates a NotifyContext but if casted can determine which signal was called  with the fired attribute.
type signalCtx struct {
	context.Context

	cancel         context.CancelFunc
	signals        []os.Signal
	ch             chan os.Signal
	Fired          os.Signal
	currConnection *net.Conn
}

func (c *signalCtx) stop() {
	c.cancel()
	signal.Stop(c.ch)
}

func (c *signalCtx) SetConnection(conn *net.Conn) {
	c.currConnection = conn
}

func NotifyContext(parent context.Context, signals ...os.Signal) (ctx context.Context, stop context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	c := &signalCtx{
		Context: ctx,
		cancel:  cancel,
		signals: signals,
	}

	c.ch = make(chan os.Signal, 1)
	signal.Notify(c.ch, c.signals...)

	if ctx.Err() == nil {
		go func() {
			select {
			case fired := <-c.ch:
				err := (*c.currConnection).Close()
				if err == nil {
					log.Info("action: close_connection | result: sucess")
				} 
				c.Fired = fired
				if fired == syscall.SIGTERM {
					log.Info("action: SIGTERM | result: success")
				} else {
					log.Info("action: SIGINT | result: success")
				}
				c.cancel()
			case <-c.Done():
			}
			log.Debug("action: stopping_monitor_goroutine | result: success")
		}()
	}
	return c, c.stop
}
