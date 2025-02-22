package widgets

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/paulmach/orb/geojson"

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

	// Matching questions
	w.content.Add(widget.NewLabel("Matching"))
	// question grid
	matchingButtonsContainer := container.NewGridWithColumns(2)

	// question buttons
	matchingButtonsContainer.Add(widget.NewButtonWithIcon("Same Bezirk", theme.Icon(theme.IconNameInfo), func() {
		err := client.AskSameBezirk(env, parentWindow)
		if err != nil {
			dialog.ShowError(err, parentWindow)
			return
		}
		log.Debug().Msg("asked same bezirk")
		refreshMap(env, parentWindow, mapWidgetPointer, historyWidgetPointer)
	}))
	matchingButtonsContainer.Add(widget.NewButtonWithIcon("Same Ortsteil", theme.Icon(theme.IconNameInfo), func() {
		err := client.AskSameOrtsteil(env, parentWindow)
		if err != nil {
			dialog.ShowError(err, parentWindow)
			return
		}
		log.Debug().Msg("asked same ortsteil")
		refreshMap(env, parentWindow, mapWidgetPointer, historyWidgetPointer)
	}))
	matchingButtonsContainer.Add(widget.NewButtonWithIcon("Ortsteil last letter", theme.Icon(theme.IconNameInfo), func() {
		err := client.AskOrtsteilLastLetter(env, parentWindow)
		if err != nil {
			dialog.ShowError(err, parentWindow)
			return
		}
		log.Debug().Msg("asked ortsteil last letter question")
		refreshMap(env, parentWindow, mapWidgetPointer, historyWidgetPointer)
	}))
	matchingButtonsContainer.Add(widget.NewButtonWithIcon("Train Service", theme.Icon(theme.IconNameInfo), func() {
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
				refreshMap(env, parentWindow, mapWidgetPointer, historyWidgetPointer)
			})
			dialogContent.Add(routeSelectionButton)
		}
		scrollableContent := container.NewVScroll(dialogContent)
		trainSelectDialog = dialog.NewCustomWithoutButtons("Select your train", scrollableContent, parentWindow)
		trainSelectDialog.Resize(fyne.NewSize(300, 600))
		trainSelectDialog.Show()
	}))
	// add matching questions container
	w.content.Add(matchingButtonsContainer)

	w.content.Add(widget.NewLabel("Relative Questions"))
	relativeButtonsContainer := container.NewGridWithColumns(2)
	relativeButtonsContainer.Add(widget.NewButtonWithIcon("McDonald's Distance", theme.Icon(theme.IconNameInfo), func() {
		err := client.AskQuestion(env, parentWindow, "closerToMcDonalds", "McDonald's Distance")
		if err != nil {
			dialog.ShowError(err, parentWindow)
		}
		refreshMap(env, parentWindow, mapWidgetPointer, historyWidgetPointer)
	}))
	relativeButtonsContainer.Add(widget.NewButtonWithIcon("IKEA Distance", theme.Icon(theme.IconNameInfo), func() {
		err := client.AskQuestion(env, parentWindow, "closerToIkea", "IKEA Distance")
		if err != nil {
			dialog.ShowError(err, parentWindow)
		}
		refreshMap(env, parentWindow, mapWidgetPointer, historyWidgetPointer)
	}))
	relativeButtonsContainer.Add(widget.NewButtonWithIcon("Spree Distance", theme.Icon(theme.IconNameInfo), func() {
		err := client.AskQuestion(env, parentWindow, "closerToSpree", "Spree Distance")
		if err != nil {
			dialog.ShowError(err, parentWindow)
		}
		refreshMap(env, parentWindow, mapWidgetPointer, historyWidgetPointer)
	}))
	w.content.Add(relativeButtonsContainer)

	w.content.Add(widget.NewLabel("Thermometer"))
	// question grid
	thermometerButtonsContainer := container.NewGridWithColumns(2)

	thermometerButtonsContainer.Add(widget.NewButtonWithIcon("start 100m Thermometer", theme.Icon(theme.IconNameInfo), func() {
		err := client.StartThermometer(env, parentWindow, 100)
		if err != nil {
			dialog.ShowError(err, parentWindow)
		}
	}))
	thermometerButtonsContainer.Add(widget.NewButtonWithIcon("end 100m Thermometer", theme.Icon(theme.IconNameInfo), func() {
		err := client.EndThermometer(env, parentWindow)
		if err != nil {
			dialog.ShowError(err, parentWindow)
			return
		}
		refreshMap(env, parentWindow, mapWidgetPointer, historyWidgetPointer)
	}))
	w.content.Add(thermometerButtonsContainer)

	// Radar questions
	w.content.Add(widget.NewLabel("Radar"))
	// question grid
	radarButtonsContainer := container.NewGridWithColumns(2)
	radarButtonsContainer.Add(widget.NewButtonWithIcon("200m Radar", theme.Icon(theme.IconNameRadioButton), func() {
		AskRadarWithRadius(env, parentWindow, 200, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButtonWithIcon("500m Radar", theme.Icon(theme.IconNameRadioButton), func() {
		AskRadarWithRadius(env, parentWindow, 500, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButtonWithIcon("1km Radar", theme.Icon(theme.IconNameRadioButton), func() {
		AskRadarWithRadius(env, parentWindow, 1000, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButtonWithIcon("2.5km Radar", theme.Icon(theme.IconNameRadioButton), func() {
		AskRadarWithRadius(env, parentWindow, 2500, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButtonWithIcon("5km Radar", theme.Icon(theme.IconNameRadioButton), func() {
		AskRadarWithRadius(env, parentWindow, 5000, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButtonWithIcon("10km Radar", theme.Icon(theme.IconNameRadioButton), func() {
		AskRadarWithRadius(env, parentWindow, 10000, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButtonWithIcon("15km Radar", theme.Icon(theme.IconNameRadioButton), func() {
		AskRadarWithRadius(env, parentWindow, 15000, mapWidgetPointer, historyWidgetPointer)
	}))
	radarButtonsContainer.Add(widget.NewButtonWithIcon("??? Radar", theme.Icon(theme.IconNameRadioButton), func() {
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
	w.content = container.NewStack(container.NewVScroll(w.content))

	return w
}

func (w *QuestionWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

func AskRadarWithRadius(env env.Env, parentWindow fyne.Window, radius float64, mapWidgetPointer *mapWidget.Map, historyWidgetPointer *HistoryWidget) {
	err := client.AskRadar(env, parentWindow, radius)
	fmt.Println("asked radar with radius", radius)
	if err != nil {
		dialog.ShowError(err, parentWindow)
		return
	}
	refreshMap(env, parentWindow, mapWidgetPointer, historyWidgetPointer)
}

func refreshMap(env env.Env, parentWindow fyne.Window, mapWidgetPointer *mapWidget.Map, historyWidgetPointer *HistoryWidget) {
	mapData, err := client.GetMapData(env, parentWindow)
	if err != nil {
		dialog.ShowError(err, parentWindow)
	}

	fc, err := geojson.UnmarshalFeatureCollection(mapData)
	if err != nil {
		dialog.ShowError(err, parentWindow)
	}

	mapWidgetPointer.SetFeatureCollection(fc)
	mapWidgetPointer.Refresh()

	// historyWidgetPointer.DoRefresh()
	historyWidgetPointer.Refresh()
}
