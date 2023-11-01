package main

import (
	"github.com/yrbb/rain"
	"github.com/yrbb/tien/manage/cmd"
)

func main() {
	app, err := rain.New()
	if err != nil {
		panic(err)
	}

	app.OnStart(func() {
		// do something
	})

	app.OnStop(func() {
		// do something
	})

	cmd.Init(app.Context(nil))

	app.Run()
}
