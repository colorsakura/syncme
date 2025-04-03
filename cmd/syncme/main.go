package main

import (
	"context"
	"sync"
	"time"

	"github.com/colorsakura/syncme/internal/discover"
	"github.com/thejerf/suture/v4"
)

type App struct {
	mainService       *suture.Supervisor
	mainServiceCancel context.CancelFunc
	stopOnce          sync.Once
	stopped           chan struct{}
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

	discoveryManager := discover.NewManager("123456", []string{})

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

func main() {
	app, err := NewApp()
	if err != nil {
		panic(err)
	}

	app.Start()
	app.Wait()
}
