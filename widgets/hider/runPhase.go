package hider

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/jkulzer/fib-client/env"
)

type HiderRunPhaseWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewHiderRunPhaseWidget(env env.Env, parentWindow fyne.Window) *HiderRunPhaseWidget {
	w := &HiderRunPhaseWidget{}
	w.ExtendBaseWidget(w)

	return w
}

func (w *HiderRunPhaseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
