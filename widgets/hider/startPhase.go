package hider

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/rs/zerolog/log"

	"fmt"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
)

type HiderStartPhaseWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewHiderStartPhaseWidget(env env.Env, parentWindow fyne.Window) *HiderStartPhaseWidget {
	w := &HiderStartPhaseWidget{}
	w.ExtendBaseWidget(w)
	w.content = container.NewVBox()

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
		fmt.Println(readySelected)
	})
	w.content.Add(readinessSelector)

	log.Info().Msg("created start phase widget")

	return w
}

func (w *HiderStartPhaseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
