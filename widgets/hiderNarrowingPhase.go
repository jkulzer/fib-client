package widgets

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/data/binding"
	// "fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// "fmt"
	// "time"
	//
	// "github.com/rs/zerolog/log"

	// "github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	// "github.com/jkulzer/fib-server/sharedModels"
)

type HiderNarrowingPhaseWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewHiderNarrowingPhaseWidget(env env.Env, parentWindow fyne.Window) *HiderNarrowingPhaseWidget {
	w := &HiderNarrowingPhaseWidget{}
	w.ExtendBaseWidget(w)

	w.content = container.NewVBox()

	return w
}

func (w *HiderNarrowingPhaseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
