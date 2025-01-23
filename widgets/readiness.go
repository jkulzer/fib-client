package widgets

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/rs/zerolog/log"

	"errors"
	"fmt"
	"time"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/helpers"

	"github.com/jkulzer/fib-server/sharedModels"
)

type ReadinessWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewReadinessWidget(env env.Env, parentWindow fyne.Window) *ReadinessWidget {
	w := &ReadinessWidget{}
	w.ExtendBaseWidget(w)
	w.content = container.NewVBox()

	appConfig, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		log.Err(err).Msg(fmt.Sprint(err))
		dialog.ShowError(err, parentWindow)
		w.content = container.NewVBox(widget.NewLabel("failed to load app config"))
		return w
	}

	readiness, err := client.IsLobbyComplete(env, parentWindow)
	if err != nil {
		errorMessage := fmt.Sprint(err)
		log.Err(err).Msg(errorMessage)
		dialog.ShowError(err, parentWindow)
		w.content.Add(widget.NewLabel(errorMessage))
		return w
	}

	if readiness {
		w.content.Add(container.NewVBox(widget.NewLabel("ready to start")))
		log.Info().Msg("lobby is ready to start")
	} else {
		w.content.Add(container.NewVBox(widget.NewLabel("Waiting for other players...")))
		log.Info().Msg("lobby not ready")
	}

	readinessSelector := widget.NewCheck("Ready", func(readySelected bool) {
		err := client.SetReadiness(env, parentWindow, readySelected)
		if err != nil {
			log.Err(err).Msg(fmt.Sprint(err))
			dialog.ShowError(err, parentWindow)
		}
	})
	w.content.Add(readinessSelector)

	log.Info().Msg("created start phase widget")

	go func() {
		for {
			timer1 := time.NewTimer(2 * time.Second)

			<-timer1.C
			readiness, err := client.IsLobbyComplete(env, parentWindow)
			if err != nil {
				errorMessage := fmt.Sprint(err)
				log.Err(err).Msg(errorMessage)
				dialog.ShowError(err, parentWindow)
				w.content.Add(widget.NewLabel(errorMessage))
				return
			}
			if readiness == true {
				switch appConfig.Role {
				case sharedModels.Hider:
					hiderRunPhaseWidget := NewHiderRunPhaseWidget(env, parentWindow)
					gameFrame := NewGameFrameWidget(env, parentWindow, hiderRunPhaseWidget)
					parentWindow.SetContent(gameFrame)
					return
				case sharedModels.Seeker:
					seekerRunPhaseWidget := NewSeekerRunPhaseWidget(env, parentWindow)
					gameFrame := NewGameFrameWidget(env, parentWindow, seekerRunPhaseWidget)
					parentWindow.SetContent(gameFrame)
					return
				default:
					message := "You are in a lobby without a valid role. Join a different one"
					err := errors.New(message)
					log.Err(err).Msg(message)
					dialog.ShowError(err, parentWindow)
					return
				}
			}
		}
	}()

	return w
}

func (w *ReadinessWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
