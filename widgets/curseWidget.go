package widgets

import (
	"reflect"
	"time"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-server/sharedModels"
)

type CurseWidget struct {
	widget.BaseWidget
	content        *widget.Accordion
	env            env.Env
	parentWindow   fyne.Window
	previousCurses []sharedModels.Card
}

func NewCurseWidget(env env.Env, parentWindow fyne.Window) *CurseWidget {
	w := &CurseWidget{
		env:            env,
		parentWindow:   parentWindow,
		previousCurses: nil,
	}
	w.ExtendBaseWidget(w)

	w.content = widget.NewAccordion()
	go func() {
		for {
			timer := time.NewTimer(2 * time.Second)
			<-timer.C

			err := w.SetContent()
			if err != nil {
				break
			}
			w.BaseWidget.Refresh()
		}
	}()

	return w
}

func (w *CurseWidget) SetContent() error {
	curses, err := client.GetCurses(w.env, w.parentWindow)
	if err != nil {
		log.Err(err).Msg("failed getting curses")
		dialog.ShowError(err, w.parentWindow)
		return err
	}

	if !reflect.DeepEqual(w.previousCurses, curses) {
		w.content.Items = nil
		for _, card := range curses {
			w.content.Append(widget.NewAccordionItem(card.Title, widget.NewLabel(card.Description)))
		}
		if len(curses) < 1 {
			w.content.Append(widget.NewAccordionItem("No active curses", container.NewVBox()))
		}
	}
	w.previousCurses = curses
	return nil
}

func (w *CurseWidget) Refresh() {
	w.SetContent()
}

func (w *CurseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewScroll(w.content))
}
