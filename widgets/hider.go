package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// "bytes"
	// "encoding/json"
	"errors"
	"fmt"
	// "net/http"
	//
	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/hiderWidget"

	"github.com/jkulzer/fib-server/sharedModels"
)

type HiderWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewHiderWidget(env env.Env, parentWindow fyne.Window) *HiderWidget {
	w := &HiderWidget{}
	w.ExtendBaseWidget(w)
	w.content = container.NewVBox()
	gamePhase := client.GetGamePhase(env, parentWindow)

	log.Info().Msg("game phase of lobby is " + fmt.Sprint(gamePhase))
	switch gamePhase {
	case sharedModels.PhaseBeforeStart:
		w.content = container.NewVBox(NewReadinessWidget(env, parentWindow))
	case sharedModels.PhaseRun:
		w.content = container.NewVBox(hiderWidget.NewRunPhaseWidget(env, parentWindow))
	case sharedModels.PhaseLocationNarrowing:
	case sharedModels.PhaseEndgame:
	case sharedModels.PhaseFinished:
	default:
		error := errors.New("invalid game state: " + fmt.Sprint(gamePhase))
		log.Err(error).Msg(fmt.Sprint(error) + " in NewHiderWidget")
		dialog.ShowError(error, parentWindow)
	}

	return w
}

func (w *HiderWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
