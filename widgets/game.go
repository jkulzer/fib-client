package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// "bytes"
	// "encoding/json"
	// "errors"
	"fmt"
	// "net/http"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/location"
	// "github.com/jkulzer/fib-client/helpers"
	"github.com/jkulzer/fib-client/models"
	// "github.com/jkulzer/fib-server/sharedModels"
)

type GameWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewGameWidget(env env.Env, parentWindow fyne.Window) *GameWidget {
	w := &GameWidget{}
	w.ExtendBaseWidget(w)

	lat, lon := location.GetLocation(parentWindow)

	logoutButton := widget.NewButton("Logout", func() {
		env.DB.Delete(&models.LoginInfo{}, 1)
		loginRegisterTabs := GetLoginRegisterTabs(env, parentWindow)
		parentWindow.SetContent(loginRegisterTabs)
	})

	leaveLobbyButton := widget.NewButton("Leave Lobby", func() {
		var loginInfo models.LoginInfo
		env.DB.First(&loginInfo)
		userName := models.LoginInfo{
			ID:         1,
			LobbyToken: "",
		}
		result := env.DB.Save(&userName)
		if result.Error != nil {
			dialog.ShowError(result.Error, parentWindow)
		} else {
			parentWindow.SetContent(NewLobbySelectionWidget(env, parentWindow))
		}
	})

	var loginInfo models.LoginInfo
	result := env.DB.First(&loginInfo)
	if result.Error != nil {
		log.Err(result.Error)
		dialog.ShowError(result.Error, parentWindow)
	}

	top := container.NewHBox(
		widget.NewLabel("Lobby code: "+loginInfo.LobbyToken),
		logoutButton,
		leaveLobbyButton,
	)

	center := container.NewVBox(
		widget.NewLabel("Latitude: "+fmt.Sprint(lat)),
		widget.NewLabel("Longitude: "+fmt.Sprint(lon)),
	)

	w.content = container.NewBorder(top, nil, nil, nil, center)

	return w
}

func (w *GameWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
