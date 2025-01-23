package hiderWidget

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

type NarrowingPhaseWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewNarrowingPhaseWidget(env env.Env, parentWindow fyne.Window) *NarrowingPhaseWidget {
	w := &NarrowingPhaseWidget{}
	w.ExtendBaseWidget(w)

	w.content = container.NewVBox()

	return w
}

func (w *NarrowingPhaseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
