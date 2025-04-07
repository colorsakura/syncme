package main

func main() {
	app, err := NewApp()
	if err != nil {
		panic(err)
	}

	if err := app.Start(); err != nil {
		panic(err)
	}
	app.Wait()
}
