package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
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
		w.content = container.NewStack(NewGameFrameWidget(env, parentWindow, center))
	} else if appConfig.Role == sharedModels.Seeker {
		center := NewSeekerWidget(env, parentWindow)
		w.content = container.NewStack(NewGameFrameWidget(env, parentWindow, center))
	} else {
		center := NewRoleSelectionWidget(env, parentWindow, appConfig.LobbyToken)
		w.content = container.NewStack(NewGameFrameWidget(env, parentWindow, center))
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
		dialog.ShowConfirm("Logout", "Are you sure you want to log out?", func(confirmed bool) {
			if confirmed {
				result := env.DB.Delete(&models.LoginInfo{}, 1)
				if result.Error != nil {
					log.Err(result.Error).Msg(fmt.Sprint(result.Error))
					dialog.ShowError(result.Error, parentWindow)
				}
				loginRegisterTabs := GetLoginRegisterTabs(env, parentWindow)
				parentWindow.SetContent(loginRegisterTabs)
			}
		}, parentWindow)
	})

	leaveLobbyButton := widget.NewButton("Leave Lobby", func() {
		confirmDialog := dialog.NewConfirm("Leave lobby", "Are you sure you want to abandon this lobby?", func(confirmed bool) {
			if confirmed {
				appConfig, err := helpers.GetAppConfig(env, parentWindow)
				if err != nil {
					log.Err(err).Msg("failed to get app config while leaving lobby")
					return
				}
				appConfig.LobbyToken = ""
				result := env.DB.Save(&appConfig)
				if result.Error != nil {
					dialog.ShowError(result.Error, parentWindow)
					return
				}
				center := NewLobbySelectionWidget(env, parentWindow)
				gameFrame := NewGameFrameWidget(env, parentWindow, center)
				parentWindow.SetContent(gameFrame)
			}

		}, parentWindow)
		confirmDialog.Show()
	})

	var loginInfo models.LoginInfo
	result := env.DB.First(&loginInfo)
	if result.Error != nil {
		log.Err(result.Error)
		dialog.ShowError(result.Error, parentWindow)
	}

	copyTokenButton := widget.NewButton("Copy code", func() {
		fyne.Clipboard.SetContent(parentWindow.Clipboard(), loginInfo.LobbyToken)
	})

	countdownText := canvas.NewText("Countdown initializing", theme.ForegroundColor())
	countdownText.Alignment = fyne.TextAlignCenter
	countdownText.TextStyle = fyne.TextStyle{Bold: true}

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond) // Smooth animation
		runStartTime, err := client.RunStartTime(env, parentWindow)
		if err != nil {
			dialog.ShowError(err, parentWindow)
		}
		gameStartTime := runStartTime.Add(sharedModels.RunDuration)

		updateText := func(s string) {
			countdownText.Text = s
			canvas.Refresh(countdownText) // Force redraw
		}

		for {
			select {
			case <-ticker.C:
				since := time.Since(gameStartTime)

				// Format with milliseconds
				since = since.Truncate(10 * time.Millisecond)
				hours := int(since.Hours())
				minutes := int(since.Minutes()) % 60
				seconds := int(since.Seconds()) % 60
				millis := since.Milliseconds() % 1000 / 10

				var timeStr string
				if hours > 0 {
					timeStr = fmt.Sprintf("%02d:%02d:%02d.%02d", hours, minutes, seconds, millis)
				} else {
					timeStr = fmt.Sprintf("%02d:%02d.%02d", minutes, seconds, millis)
				}

				updateText(timeStr)
			}
		}
	}()

	top := container.NewHBox(
		widget.NewLabel("Lobby code: "+loginInfo.LobbyToken),
		copyTokenButton,
		logoutButton,
		leaveLobbyButton,
		countdownText,
	)

	w.content = container.NewBorder(top, nil, nil, nil, center)

	return w
}

func (w *GameFrameWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
