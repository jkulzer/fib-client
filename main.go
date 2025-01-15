package main

import (
	"fyne.io/fyne/v2/app"

	// "fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/widget"

	"github.com/jkulzer/fib-client/db"
	"github.com/jkulzer/fib-client/models"
	"github.com/jkulzer/fib-client/widgets"

	"github.com/rs/zerolog/log"
)

func main() {
	app := app.NewWithID("dev.jkulzer.findinberlin")
	w := app.NewWindow("FindInBerlin")

	env := db.InitDB(app)

	env.Url = "http://192.168.69.235:3001"
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
	// name := widget.NewLabel("(tap lookup)")
	// w.SetContent(container.NewVBox(name,
	// 	widget.NewButton("Lookup", func() {
	// 		name.SetText("test")
	// 	})))

	w.ShowAndRun()
}
