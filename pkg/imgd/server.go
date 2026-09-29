package imgd

import (
	"context"
	"errors"
	"flag"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// ServerConfig contains only the settings used by the HTTP application.
type ServerConfig struct {
	HTTPListenAddress             string
	HTTPListenPort                int
	RegisterInstrumentation       bool
	ServerGracefulShutdownTimeout time.Duration
	HTTPServerReadTimeout         time.Duration
	HTTPServerWriteTimeout        time.Duration
	HTTPServerIdleTimeout         time.Duration
}

func (c *ServerConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.HTTPListenAddress, "server.http-listen-address", "", "HTTP server listen address")
	f.IntVar(&c.HTTPListenPort, "server.http-listen-port", 80, "HTTP server listen port")
	f.BoolVar(&c.RegisterInstrumentation, "server.register-instrumentation", true, "Serve metrics and pprof endpoints")
	f.DurationVar(&c.ServerGracefulShutdownTimeout, "server.graceful-shutdown-timeout", 30*time.Second, "Timeout for graceful shutdown")
	f.DurationVar(&c.HTTPServerReadTimeout, "server.http-read-timeout", 30*time.Second, "HTTP read timeout")
	f.DurationVar(&c.HTTPServerWriteTimeout, "server.http-write-timeout", 30*time.Second, "HTTP write timeout")
	f.DurationVar(&c.HTTPServerIdleTimeout, "server.http-idle-timeout", 120*time.Second, "HTTP idle timeout")
}

func (i *Imgd) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return i.RunContext(ctx)
}

// RunContext serves HTTP until cancellation, then drains requests and closes caches.
func (i *Imgd) RunContext(ctx context.Context) error {
	defer i.Close()
	cfg := i.Cfg.Server
	server := &http.Server{
		Addr:         net.JoinHostPort(cfg.HTTPListenAddress, strconv.Itoa(cfg.HTTPListenPort)),
		Handler:      i.Handler(),
		ReadTimeout:  cfg.HTTPServerReadTimeout,
		WriteTimeout: cfg.HTTPServerWriteTimeout,
		IdleTimeout:  cfg.HTTPServerIdleTimeout,
	}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	i.Cfg.Logger.Infof("Listening on %s", listener.Addr())
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ServerGracefulShutdownTimeout)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			server.Close()
		}
		<-done
		return err
	}
}
