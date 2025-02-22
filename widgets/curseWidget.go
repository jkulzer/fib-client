package widgets

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// "github.com/rs/zerolog/log"

	// "github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	// "github.com/jkulzer/fib-server/sharedModels"
)

type CurseWidget struct {
	widget.BaseWidget
	content      *widget.Accordion
	env          env.Env
	parentWindow fyne.Window
}

func NewCurseWidget(env env.Env, parentWindow fyne.Window) *CurseWidget {
	w := &CurseWidget{}
	w.ExtendBaseWidget(w)
	w.content = widget.NewAccordion()
	return w
}

func (w *CurseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewScroll(w.content))
}
