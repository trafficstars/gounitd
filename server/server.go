package server

import (
	"errors"
	"github.com/trafficstars/metrics"
	"sync"
	"time"

	"github.com/trafficstars/fasthttp"
)

type serverState int32

const (
	serverStateStopped = iota
	serverStateStarting
	serverStateRunning
	serverStateStopping
)

var (
	ErrAlreadyStarted = errors.New(`[gounit-server] already started`)
)

type Server struct {
	locker sync.Mutex

	State     serverState
	Frontends []*Frontend
	Backends  []*Backend

	ErrorLogger  Logger
	AccessLogger Logger
}

func NewServer(cfg *Config) (*Server, error) {
	srv := &Server{}
	for _, frontendCfg := range cfg.Frontends {
		frontend, err := newFrontend(srv, frontendCfg)
		if err != nil {
			return nil, err
		}
		srv.Frontends = append(srv.Frontends, frontend)
	}
	for _, backendCfg := range cfg.Backends {
		backend, err := newBackend(srv, backendCfg)
		if err != nil {
			return nil, err
		}
		srv.Backends = append(srv.Backends, backend)
	}
	return srv, nil
}

func (srv *Server) Start() error {
	srv.locker.Lock()
	defer srv.locker.Unlock()
	oldState := srv.setState(serverStateStarting)
	if oldState != serverStateStopped {
		srv.setStateFrom(oldState, serverStateStarting)
		return ErrAlreadyStarted
	}
	for _, frontend := range srv.Frontends {
		err := frontend.Start()
		if err != nil {
			srv.setState(serverStateStopped)
			return err
		}
	}
	for _, backend := range srv.Backends {
		err := backend.Start()
		if err != nil {
			srv.setState(serverStateStopped)
			return err
		}
	}
	srv.setState(serverStateRunning)
	return nil
}

func (srv *Server) GetState() serverState {
	//return serverState(atomic.LoadUint32((*uint32)((unsafe.Pointer)(&srv.State))))
	srv.locker.Lock()
	defer srv.locker.Unlock()
	return srv.State
}

func (srv *Server) setState(newState serverState) serverState {
	//return serverState(atomic.SwapUint32((*uint32)((unsafe.Pointer)(&srv.State)), uint32(newState)))
	var oldState serverState
	oldState, srv.State = srv.State, newState
	return oldState
}

func (srv *Server) setStateFrom(newState, oldState serverState) {
	/*for !atomic.CompareAndSwapUint32((*uint32)((unsafe.Pointer)(&srv.State)), uint32(oldState), uint32(newState)) {
		time.Sleep(time.Microsecond * time.Duration(rand.Intn(1000)))
	}*/
	if srv.State != oldState {
		panic(`should not happened`)
	}
	srv.State = newState
}

func (srv *Server) Wait() error {
	for srv.GetState() != serverStateStopped {
		time.Sleep(time.Second)
	}
	return nil
}

func (srv *Server) metricsConsider(startTime time.Time, f *Frontend, b *Backend, ctx *fasthttp.RequestCtx) {
	var backendAddress string
	if b != nil {
		backendAddress = b.Address
	}
	tags := metrics.NewFastTags().
		Set(`frontend`, f.ListenFamily+":"+f.ListenAddress).
		Set(`backend`, backendAddress).
		Set(`host`, string(ctx.Request.Host())).
		Set(`code`, ctx.Response.StatusCode())
	metrics.Count(`requests`, tags).Increment()
	metrics.TimingBuffered(`request_latency`, tags).ConsiderValue(time.Since(startTime))
	tags.Release()
}

func (srv *Server) HandleRequest(f *Frontend, ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	for _, backend := range srv.Backends {
		if backend.IsFits(ctx) {
			backend.HandleRequest(f, ctx)
			srv.metricsConsider(startTime, f, backend, ctx)
			return
		}
	}
	srv.send404(ctx)
	srv.metricsConsider(startTime, f, nil, ctx)
}

func (srv *Server) send404(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(404)
}
