package widgets

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/location"
	"github.com/jkulzer/fib-client/mapWidget"
)

type QuestionWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewQuestionWidget(env env.Env, parentWindow fyne.Window, mapWidgetPointer *mapWidget.Map, historyWidgetPointer *HistoryWidget) *QuestionWidget {
	w := &QuestionWidget{}
	w.ExtendBaseWidget(w)
	w.content = container.NewVBox()

	setLocationButton := widget.NewButton("Set Location", func() {
		go func() {

			locationPoint, err := location.GetLocation(parentWindow)
			if err != nil {
				dialog.ShowError(err, parentWindow)
				return
			}
			err = client.SaveLocation(env, parentWindow, locationPoint)
			if err != nil {
				dialog.ShowError(err, parentWindow)
			}
		}()
	})
	w.content.Add(setLocationButton)
	var questionHeaderSize float32 = 18.0

	matchingText := canvas.NewText("Ja/Nein Fragen:", theme.Color(theme.ColorNameForeground))
	matchingText.TextSize = questionHeaderSize // Big font size
	matchingText.TextStyle = fyne.TextStyle{Bold: true}
	w.content.Add(matchingText)
	// question grid
	matchingButtonsContainer := container.NewGridWithColumns(2)

	// question buttons
	buttonName := "Selber Bezirk?"
	matchingButtonsContainer.Add(widget.NewButton(buttonName, func() {
		dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question "+buttonName, func(confirmed bool) {
			if confirmed {
				err := client.AskSameBezirk(env, parentWindow)
				if err != nil {
					dialog.ShowError(err, parentWindow)
					return
				}
				log.Debug().Msg("asked same bezirk")
				refreshMap(mapWidgetPointer, historyWidgetPointer)
			}
		}, parentWindow)
	}))
	buttonName = "Selber Ortsteil?"
	matchingButtonsContainer.Add(widget.NewButton(buttonName, func() {
		dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question "+buttonName, func(confirmed bool) {
			if confirmed {
				err := client.AskSameOrtsteil(env, parentWindow)
				if err != nil {
					dialog.ShowError(err, parentWindow)
					return
				}
				log.Debug().Msg("asked same ortsteil")
				refreshMap(mapWidgetPointer, historyWidgetPointer)
			}
		}, parentWindow)
	}))
	buttonName = "Selber letzter Buchstabe des Ortsteils?"
	matchingButtonsContainer.Add(widget.NewButton(buttonName, func() {
		dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question "+buttonName, func(confirmed bool) {
			if confirmed {
				err := client.AskOrtsteilLastLetter(env, parentWindow)
				if err != nil {
					dialog.ShowError(err, parentWindow)
					return
				}
				log.Debug().Msg("asked ortsteil last letter question")
				refreshMap(mapWidgetPointer, historyWidgetPointer)
			}
		}, parentWindow)
	}))
	buttonName = "Hält der Zug in der Nähe des Hiders?"
	matchingButtonsContainer.Add(widget.NewButton(buttonName, func() {
		dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question "+buttonName, func(confirmed bool) {
			if confirmed {
				log.Info().Msg("asked train service question")
				closeRouteList, err := client.GetCloseRoutes(env, parentWindow)
				if err != nil {
					log.Err(err).Msg("failed asking train service question")
					dialog.ShowError(err, parentWindow)
					return
				}
				var trainSelectDialog *dialog.CustomDialog
				dialogContent := container.NewGridWithColumns(1)
				for _, route := range closeRouteList.Routes {
					routeSelectionButton := widget.NewButton(route.Name, func() {
						log.Info().Msg("selected route " + route.Name + " with ID " + fmt.Sprint(route.RouteID))
						err := client.AskTrainservice(env, parentWindow, route.RouteID)
						if err != nil {
							log.Err(err).Msg("failed asking train service question")
							dialog.ShowError(err, parentWindow)
							return
						}
						trainSelectDialog.Hide()
						refreshMap(mapWidgetPointer, historyWidgetPointer)
					})
					dialogContent.Add(routeSelectionButton)
				}
				if len(closeRouteList.Routes) <= 0 {
					dialogContent.Add(widget.NewLabel("Not on a train line"))
				}
				scrollableContent := container.NewVScroll(dialogContent)
				trainSelectDialog = dialog.NewCustom("Select your train", "dismiss", scrollableContent, parentWindow)
				trainSelectDialog.Resize(fyne.NewSize(300, 600))
				trainSelectDialog.Show()
			}
		}, parentWindow)
	}))
	// add matching questions container
	w.content.Add(matchingButtonsContainer)

	relativeText := canvas.NewText("Relative Fragen:", theme.Color(theme.ColorNameForeground))
	relativeText.TextSize = 18 // Big font size
	relativeText.TextStyle = fyne.TextStyle{Bold: true}
	w.content.Add(relativeText)
	w.content.Add(widget.NewLabel("Näher oder weiter weg von..."))
	relativeButtonsContainer := container.NewGridWithColumns(2)
	buttonName = "...einem McDonald's?"
	relativeButtonsContainer.Add(widget.NewButton(buttonName, func() {
		dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question?", func(confirmed bool) {
			if confirmed {
				err := client.AskQuestion(env, parentWindow, "closerToMcDonalds", "McDonald's Distance")
				if err != nil {
					dialog.ShowError(err, parentWindow)
				}
				refreshMap(mapWidgetPointer, historyWidgetPointer)
			}
		}, parentWindow)
	}))
	buttonName = "...einem IKEA?"
	relativeButtonsContainer.Add(widget.NewButton(buttonName, func() {
		dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question?", func(confirmed bool) {
			if confirmed {
				err := client.AskQuestion(env, parentWindow, "closerToIkea", "IKEA Distance")
				if err != nil {
					dialog.ShowError(err, parentWindow)
				}
				refreshMap(mapWidgetPointer, historyWidgetPointer)
			}
		}, parentWindow)
	}))
	relativeButtonsContainer.Add(widget.NewButton("...der Spree?", func() {
		dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question?", func(confirmed bool) {
			if confirmed {
				err := client.AskQuestion(env, parentWindow, "closerToSpree", "Spree Distance")
				if err != nil {
					dialog.ShowError(err, parentWindow)
				}
				refreshMap(mapWidgetPointer, historyWidgetPointer)
			}
		}, parentWindow)
	}))
	w.content.Add(relativeButtonsContainer)

	thermometerText := canvas.NewText("Thermometer:", theme.Color(theme.ColorNameForeground))
	thermometerText.TextSize = 18 // Big font size
	thermometerText.TextStyle = fyne.TextStyle{Bold: true}
	w.content.Add(thermometerText)
	// question grid
	thermometerButtonsContainer := container.NewGridWithColumns(2)

	thermometerButtonsContainer.Add(widget.NewButton("Starte Thermometer", func() {
		dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question?", func(confirmed bool) {
			if confirmed {
				err := client.StartThermometer(env, parentWindow, 100)
				if err != nil {
					dialog.ShowError(err, parentWindow)
				}
			}
		}, parentWindow)
	}))
	thermometerButtonsContainer.Add(widget.NewButton("Ende Thermometer", func() {
		dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question?", func(confirmed bool) {
			if confirmed {
				err := client.EndThermometer(env, parentWindow)
				if err != nil {
					dialog.ShowError(err, parentWindow)
					return
				}
				refreshMap(mapWidgetPointer, historyWidgetPointer)
			}
		}, parentWindow)
	}))
	w.content.Add(thermometerButtonsContainer)

	// Radar questions
	radarText := canvas.NewText("Radar:", theme.Color(theme.ColorNameForeground))
	radarText.TextSize = 18 // Big font size
	radarText.TextStyle = fyne.TextStyle{Bold: true}
	w.content.Add(radarText)
	// question grid
	radarButtonsContainer := container.NewGridWithColumns(2)
	radarButtonsContainer.Add(widget.NewButton("200m Radar", func() {
		AskRadarWithRadius(env, parentWindow, 200, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButton("500m Radar", func() {
		AskRadarWithRadius(env, parentWindow, 500, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButton("1km Radar", func() {
		AskRadarWithRadius(env, parentWindow, 1000, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButton("2.5km Radar", func() {
		AskRadarWithRadius(env, parentWindow, 2500, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButton("5km Radar", func() {
		AskRadarWithRadius(env, parentWindow, 5000, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButton("10km Radar", func() {
		AskRadarWithRadius(env, parentWindow, 10000, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButton("15km Radar", func() {
		AskRadarWithRadius(env, parentWindow, 15000, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButton("??? Radar", func() {
		// TODO
		radiusEntry := widget.NewEntry()
		formItems := []*widget.FormItem{
			{Text: "Radius", Widget: radiusEntry},
		}
		callback := func(boolean bool) {
			if boolean {
				radiusFloat, err := strconv.ParseFloat(radiusEntry.Text, 64)
				if err != nil {
					dialog.ShowError(err, parentWindow)
					return
				}
				AskRadarWithRadius(env, parentWindow, radiusFloat, mapWidgetPointer, historyWidgetPointer)
			}
		}
		formDialog := dialog.NewForm("Enter radar radius in meters:", "Confirm", "Dismiss", formItems, callback, parentWindow)
		formDialog.Show()
	}))
	// add radar questions container
	w.content.Add(radarButtonsContainer)

	endgameQuestionsText := canvas.NewText("Endgame questions:", theme.Color(theme.ColorNameForeground))
	endgameQuestionsText.TextSize = 18 // Big font size
	endgameQuestionsText.TextStyle = fyne.TextStyle{Bold: true}
	w.content.Add(endgameQuestionsText)
	endgameQuestionsContainer := container.NewGridWithColumns(2)
	endgameQuestionsContainer.Add(widget.NewButton("Hiding zone", func() {
		dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question?", func(confirmed bool) {
			if confirmed {
				err := client.AskQuestion(env, parentWindow, "isInHidingZone", "In hiding zone")
				if err != nil {
					dialog.ShowError(err, parentWindow)
				}
				refreshMap(mapWidgetPointer, historyWidgetPointer)
			}
		}, parentWindow)
	}))
	w.content.Add(endgameQuestionsContainer)

	w.content = container.NewStack(container.NewVScroll(w.content))
	return w
}

func (w *QuestionWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

func AskRadarWithRadius(env env.Env, parentWindow fyne.Window, radius float64, mapWidgetPointer *mapWidget.Map, historyWidgetPointer *HistoryWidget) {
	dialog.ShowConfirm("Ask question", "Are you sure you want to ask the question?", func(confirmed bool) {
		if confirmed {
			err := client.AskRadar(env, parentWindow, radius)
			fmt.Println("asked radar with radius", radius)
			if err != nil {
				dialog.ShowError(err, parentWindow)
				return
			}
			refreshMap(mapWidgetPointer, historyWidgetPointer)
		}
	}, parentWindow)
}

func refreshMap(mapWidgetPointer *mapWidget.Map, historyWidgetPointer *HistoryWidget) {
	mapWidgetPointer.Refresh()

	historyWidgetPointer.Refresh()
}
