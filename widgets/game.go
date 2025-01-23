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
	// "github.com/jkulzer/fib-client/location"
	"github.com/jkulzer/fib-client/helpers"
	"github.com/jkulzer/fib-client/models"
	"github.com/jkulzer/fib-server/sharedModels"
)

type GameWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewGameWidget(env env.Env, parentWindow fyne.Window) *GameWidget {

	w := &GameWidget{}
	w.ExtendBaseWidget(w)

	appConfig, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		log.Err(err).Msg("failed to get app config in game widget")
	}

	if appConfig.Role == sharedModels.Hider {
		log.Info().Msg("Found hider role in database")
		center := NewHiderWidget(env, parentWindow)
		w.content = container.NewVBox(NewGameFrameWidget(env, parentWindow, center))
	} else if appConfig.Role == sharedModels.Seeker {
		center := NewSeekerWidget(env, parentWindow)
		w.content = container.NewVBox(NewGameFrameWidget(env, parentWindow, center))
	} else {
		center := NewRoleSelectionWidget(env, parentWindow, appConfig.LobbyToken)
		w.content = container.NewVBox(NewGameFrameWidget(env, parentWindow, center))
	}

	return w
}

func (w *GameWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

type GameFrameWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewGameFrameWidget(env env.Env, parentWindow fyne.Window, center fyne.CanvasObject) *GameFrameWidget {

	w := &GameFrameWidget{}
	w.ExtendBaseWidget(w)

	// lat, lon := location.GetLocation(parentWindow)

	logoutButton := widget.NewButton("Logout", func() {
		result := env.DB.Delete(&models.LoginInfo{}, 1)
		if result.Error != nil {
			log.Err(result.Error).Msg(fmt.Sprint(result.Error))
			dialog.ShowError(result.Error, parentWindow)
		}
		loginRegisterTabs := GetLoginRegisterTabs(env, parentWindow)
		parentWindow.SetContent(loginRegisterTabs)
	})

	leaveLobbyButton := widget.NewButton("Leave Lobby", func() {
		appConfig, err := helpers.GetAppConfig(env, parentWindow)
		if err != nil {
			log.Err(err).Msg("failed to get app config while leaving lobby")
		} else {
			appConfig.LobbyToken = ""
			result := env.DB.Save(&appConfig)
			if result.Error != nil {
				dialog.ShowError(result.Error, parentWindow)
			} else {
				parentWindow.SetContent(NewLobbySelectionWidget(env, parentWindow))
			}
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

	w.content = container.NewBorder(top, nil, nil, nil, center)

	return w
}

func (w *GameFrameWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
