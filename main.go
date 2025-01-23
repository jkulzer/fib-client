package main

import (
	"fyne.io/fyne/v2/app"

	// "fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/widget"

	"github.com/jkulzer/fib-client/db"
	"github.com/jkulzer/fib-client/location"
	"github.com/jkulzer/fib-client/models"
	"github.com/jkulzer/fib-client/widgets"

	"github.com/rs/zerolog/log"

	"fmt"
	"os"
)

func main() {
	app := app.NewWithID("dev.jkulzer.findinberlin")
	w := app.NewWindow("FindInBerlin")

	var dbSubpath string
	if len(os.Args) >= 2 {
		log.Info().Msg("db subpath: " + os.Args[1])
		dbSubpath = os.Args[1]
	} else {
		dbSubpath = "sqlite"
	}

	env := db.InitDB(app, dbSubpath)

	env.Url = "http://localhost:3001"
	// env.Url = "http://192.168.69.235:3001"
	// env.Url = "http://192.168.178.178:3001"
	var loginInfo models.LoginInfo
	result := env.DB.First(&loginInfo)
	if result.Error != nil {
		log.Warn().Msg("couldn't find token, starting login sequence")
		loginRegisterTabs := widgets.GetLoginRegisterTabs(env, w)
		w.SetContent(loginRegisterTabs)
	} else {
		if loginInfo.LobbyToken != "" {
			w.SetContent(widgets.NewGameWidget(env, w))
		} else {
			w.SetContent(widgets.NewLobbyWidget(env, w))
		}
	}

	go func() {
		point, _ := location.GetLocation(w)
		fmt.Println(point)
	}()

	w.ShowAndRun()
}
