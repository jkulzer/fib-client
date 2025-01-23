package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// "bytes"
	// "encoding/json"
	"errors"
	"fmt"
	// "net/http"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	// "github.com/jkulzer/fib-client/models"
	"github.com/jkulzer/fib-server/sharedModels"
)

type SeekerWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewSeekerWidget(env env.Env, parentWindow fyne.Window) *SeekerWidget {
	w := &SeekerWidget{}
	w.ExtendBaseWidget(w)
	w.content = container.NewVBox()

	gamePhase := client.GetGamePhase(env, parentWindow)

	switch gamePhase {
	case sharedModels.PhaseBeforeStart:
		w.content = container.NewVBox(NewReadinessWidget(env, parentWindow))
	case sharedModels.PhaseRun:
		w.content = container.NewVBox(NewSeekerRunPhaseWidget(env, parentWindow))
	case sharedModels.PhaseLocationNarrowing:
	case sharedModels.PhaseEndgame:
	case sharedModels.PhaseFinished:
	default:
		error := errors.New("invalid game state: " + fmt.Sprint(gamePhase))
		log.Err(error).Msg(fmt.Sprint(error) + " in NewSeekerWidget")
		dialog.ShowError(error, parentWindow)
	}

	return w
}

func (w *SeekerWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
