package widgets

import (
	"errors"
	"fmt"
	"slices"
	"time"

	fyne "fyne.io/fyne/v2"
	// "fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-server/sharedModels"
)

type CardsWidget struct {
	widget.BaseWidget
	content      *fyne.Container
	env          env.Env
	parentWindow fyne.Window
}

func NewCardsWidget(env env.Env, parentWindow fyne.Window) *CardsWidget {
	w := &CardsWidget{}
	w.ExtendBaseWidget(w)
	w.env = env
	w.parentWindow = parentWindow
	w.content = container.NewVBox()

	go func() {
		for {
			timer := time.NewTimer(2 * time.Second)
			<-timer.C

			w.Refresh()
		}
	}()
	return w
}

func (w *CardsWidget) CreateRenderer() fyne.WidgetRenderer {
	cardActions, err := client.GetCardActions(w.env, w.parentWindow)
	if err != nil {
		dialog.ShowError(err, w.parentWindow)
	}
	w.content.Add(
		widget.NewButton("Check for in progress draw", func() {
			draw, err := client.GetDraw(w.env, w.parentWindow)
			if err != nil {
				log.Err(err).Msg("failed getting draw")
				dialog.ShowError(err, w.parentWindow)
				return
			}
			if len(draw.Cards) < 1 {
				dialog.ShowInformation("Card draw", "There is no draw in progress", w.parentWindow)
			} else {
				fmt.Println(draw.ToPick)
				cardSelectDialog(w, w.env, w.parentWindow, draw.ToPick)
			}
		}),
	)
	if len(cardActions.Draws) == 0 {
		w.content.Add(widget.NewLabel("no remaining draws"))
	} else {
		drawContainer := container.NewVBox()
		for _, cardAction := range cardActions.Draws {
			drawContainer.Add(
				container.NewVBox(
					widget.NewLabel("Draw "+fmt.Sprint(cardAction.CardsToDraw)+" cards and pick "+fmt.Sprint(cardAction.CardsToPick)),
					widget.NewButton("Draw!", func() {
						log.Debug().Msg("use card draw with ID " + fmt.Sprint(cardAction.DrawID))
						err := client.DrawCards(w.env, w.parentWindow, cardAction.DrawID)
						if errors.Is(err, client.ErrAlreadyDrewCards) {
							dialog.ShowInformation("Card drawing", "You need to pick cards from your previous draw before you can draw new cards", w.parentWindow)
							return
						}
						if err != nil {
							log.Err(err).Msg("failed drawing cards")
							dialog.ShowError(err, w.parentWindow)
							return
						}
						cardSelectDialog(w, w.env, w.parentWindow, cardAction.CardsToPick)
					}),
				),
			)
		}
		w.content.Add(drawContainer)
	}
	return widget.NewSimpleRenderer(container.NewScroll(w.content))
}

func cardSelectDialog(w *CardsWidget, env env.Env, parentWindow fyne.Window, maxCards uint) {
	drawnCardsContainer := NewCardSelectWidget(env, parentWindow, maxCards)
	cardDrawDialog := dialog.NewCustom("Drawn cards", "dismiss", drawnCardsContainer, w.parentWindow)
	cardDrawDialog.Resize(fyne.NewSize(300, 600))
	cardDrawDialog.Show()
}

type CardWidget struct {
	widget.BaseWidget
	content          *fyne.Container
	selected         bool
	selectedText     *widget.Label
	widgetType       CardWidgetType
	cardSelectWidget *CardSelectWidget
	card             sharedModels.Card
}

type CardWidgetType int

const (
	DrawCardWidget CardWidgetType = iota
	PlayCardWidget
)

func NewCardWidget(card sharedModels.Card, widgetType CardWidgetType, cardSelectWidget *CardSelectWidget) *CardWidget {
	w := &CardWidget{}
	w.ExtendBaseWidget(w)
	title := widget.NewLabel(card.Title)
	title.TextStyle = fyne.TextStyle{
		Bold:      true,
		Underline: true,
	}
	w.card = card
	w.cardSelectWidget = cardSelectWidget
	w.widgetType = widgetType
	// title.Wrapping = fyne.TextWrapWord
	var description *widget.Label
	if w.card.Description != "" {
		description = widget.NewLabel(w.card.Description)
	} else {
		description = widget.NewLabel("No Description")
	}

	description.Wrapping = fyne.TextWrapWord
	w.content = container.NewVBox(title, description)
	w.selectedText = widget.NewLabel("selected")

	return w
}

func (w *CardWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

func (w *CardWidget) Refresh() {
	w.BaseWidget.Refresh()
}

func (w *CardWidget) Tapped(*fyne.PointEvent) {
	switch w.widgetType {
	case DrawCardWidget:
		if w.selected {
			w.selected = false
			w.content.Remove(w.selectedText)
			w.cardSelectWidget.RemoveCardFromPool(w.card)
		} else {
			w.selected = true
			w.content.Add(w.selectedText)
			w.cardSelectWidget.AddCardToPool(w.card)
		}
	}
	w.cardSelectWidget.Refresh()
}

type CardSelectWidget struct {
	widget.BaseWidget
	content      *fyne.Container
	env          env.Env
	parentWindow fyne.Window
	cards        []sharedModels.Card
	maxCards     uint
}

func NewCardSelectWidget(env env.Env, parentWindow fyne.Window, maxCards uint) *CardSelectWidget {
	w := &CardSelectWidget{
		content:      container.NewGridWithRows(2),
		maxCards:     maxCards,
		env:          env,
		parentWindow: parentWindow,
	}
	w.ExtendBaseWidget(w)
	draw, err := client.GetDraw(w.env, w.parentWindow)
	if err != nil {
		log.Err(err).Msg("failed getting draw")
		dialog.ShowError(err, w.parentWindow)
		return w
	}
	for _, card := range draw.Cards {
		w.content.Add(NewCardWidget(card, DrawCardWidget, w))
	}
	return w
}

func (w *CardSelectWidget) Refresh() {
	log.Debug().Msg("amount of cards: " + fmt.Sprint(len(w.cards)) + " with a limit of " + fmt.Sprint(w.maxCards))
	if len(w.cards) <= int(w.maxCards) && len(w.cards) > 0 {
		log.Debug().Msg("valid amount of cards to pick selected")
		w.content = container.NewBorder(
			nil,
			// bottom
			widget.NewButton("Pick cards", func() {
			}),
			nil,
			nil,
			// center
			w.content,
		)
	}
}

func (w *CardSelectWidget) CreateRenderer() fyne.WidgetRenderer {

	return widget.NewSimpleRenderer(w.content)

}

func (w *CardSelectWidget) AddCardToPool(card sharedModels.Card) {
	w.cards = append(w.cards, card)
}
func (w *CardSelectWidget) RemoveCardFromPool(cardToDelete sharedModels.Card) {
	for index, card := range w.cards {
		if card == cardToDelete {
			log.Debug().Msg("deleting card from pool")
			w.cards = slices.Delete(w.cards, index, 1)
			return
		}
	}
}
