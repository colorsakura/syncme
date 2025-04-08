package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/colorsakura/syncme/internal/discover"
	"github.com/colorsakura/syncme/internal/protocol"
	"github.com/thejerf/suture/v4"
)

type App struct {
	mainService       *suture.Supervisor
	mainServiceCancel context.CancelFunc
	stopOnce          sync.Once
	stopped           chan struct{}
}

type lateAddressLister struct {
	discover.AddressLister
}

func NewApp() (*App, error) {
	app := &App{
		stopped: make(chan struct{}),
	}
	close(app.stopped) // Hasn't been started, so shouldn't block on Wait.
	return app, nil
}

func (app *App) Start() error {
	app.mainService = suture.New("main", suture.Spec{})

	app.stopped = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	errChan := app.mainService.ServeBackground(ctx)

	app.mainServiceCancel = cancel

	go app.wait(errChan)

	addrLister := &lateAddressLister{}

	uid, _ := protocol.NewDeviceID([]byte{})

	discoveryManager := discover.NewManager(uid, []string{}, addrLister, log.Default())

	app.mainService.Add(discoveryManager)

	return nil
}

func (a *App) wait(errChan <-chan error) {
	<-errChan

	done := make(chan struct{})
	go func() {
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
	}

	close(a.stopped)
}

func (app *App) Wait() {
	<-app.stopped
}
