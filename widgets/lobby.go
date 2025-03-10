package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/helpers"
	"github.com/jkulzer/fib-client/models"
	"github.com/jkulzer/fib-server/sharedModels"
)

type LobbyWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewLobbyWidget(env env.Env, parentWindow fyne.Window) *LobbyWidget {
	w := &LobbyWidget{}
	w.ExtendBaseWidget(w)

	logoutButton := widget.NewButton("Logout", func() {
		dialog.ShowConfirm("Logout", "Are you sure you want to log out?", func(confirmed bool) {
			if confirmed {
				result := env.DB.Delete(&models.LoginInfo{}, 1)
				if result.Error != nil {
					log.Err(result.Error).Msg(fmt.Sprint(result.Error))
					dialog.ShowError(result.Error, parentWindow)
				}
				loginRegisterTabs := GetLoginRegisterTabs(env, parentWindow)
				parentWindow.SetContent(loginRegisterTabs)
			}
		}, parentWindow)
	})

	top := container.NewHBox(logoutButton)

	middle := NewLobbySelectionWidget(env, parentWindow)

	w.content = container.NewBorder(top, nil, nil, nil, middle)

	return w
}

func (w *LobbyWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

type LobbySelectionWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewLobbySelectionWidget(env env.Env, parentWindow fyne.Window) *LobbySelectionWidget {
	w := &LobbySelectionWidget{}
	w.ExtendBaseWidget(w)
	lobbyCodeEntry := widget.NewEntry()
	lobbyCodeEntry.SetPlaceHolder("AG5L3T")
	lobbyCodeEntry.Validator = validation.NewRegexp(sharedModels.LobbyCodeRegex, "Lobby code must be 6 characters")

	var appConfiguration models.LoginInfo
	env.DB.First(&appConfiguration)

	lobbyEntryForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Lobby Code", Widget: lobbyCodeEntry},
		},
		OnSubmit: func() {
			go func() {
				joinLobby(lobbyCodeEntry.Text, parentWindow, env)
			}()
		},
		SubmitText: "Join Lobby",
	}

	lobbyCreationButton := widget.NewButton("Create Lobby", func() {
		req, err := http.NewRequest("POST", env.Url+"/lobby/create", nil)
		if err != nil {
			dialog.ShowError(err, parentWindow)
		}
		// error handing already handled since it shows a popup
		loginInfo, err := helpers.GetAppConfig(env, parentWindow)
		if err == nil {
			req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Warn().Msg("couldn't make request" + fmt.Sprint(err))
				dialog.ShowError(err, parentWindow)
				return
			}
			switch res.StatusCode {
			case http.StatusCreated:
				var responseStruct sharedModels.LobbyCreationResponse
				responseBytes, err := helpers.ReadHttpResponse(res.Body)
				if err != nil {
					log.Err(err).Msg("failed to read http response for lobby creation")
					dialog.ShowError(err, parentWindow)
				}
				err = json.Unmarshal(responseBytes, &responseStruct)
				if err != nil {
					message := "Failed to unmarshal http response"
					log.Err(err).Msg(message)
					dialog.ShowError(err, parentWindow)
				}
				appConfig, err := helpers.GetAppConfig(env, parentWindow)
				if err != nil {
					log.Err(err)
					dialog.ShowError(err, parentWindow)
				}
				appConfig.LobbyToken = responseStruct.LobbyToken
				// tries to create the user in the db
				result := env.DB.Save(&appConfig)
				if result.Error != nil {
					log.Err(err).Msg("failed to save configuration in database")
					dialog.ShowError(err, parentWindow)
				} else {
					log.Info().Msg("created lobby " + responseStruct.LobbyToken)
					creationDialog := dialog.NewCustom("Lobby Creation", "Close", container.NewVBox(
						widget.NewLabel("Created lobby with token \""+responseStruct.LobbyToken+"\""),
						widget.NewButton("Copy to clipboard", func() {
							fyne.Clipboard.SetContent(parentWindow.Clipboard(), responseStruct.LobbyToken)
						}),
					),
						parentWindow,
					)
					creationDialog.Show()

					// dialog.ShowInformation("Lobby Creation", "Created lobby with token \""+responseStruct.LobbyToken+"\"", parentWindow)
					joinLobby(responseStruct.LobbyToken, parentWindow, env)
				}

			case http.StatusForbidden:
				message := "Not logged in. Log out and back in."
				log.Info().Msg(message)
				error := errors.New(message)
				dialog.ShowError(error, parentWindow)
			default:
				dialog.ShowError(errors.New(fmt.Sprint(res.StatusCode)), parentWindow)
			}
		}
	})

	lobbyJoin := container.NewVBox(lobbyEntryForm)
	lobbyCreate := container.NewVBox(lobbyCreationButton)

	w.content = container.NewVBox(
		widget.NewLabel("Join a Lobby"),
		lobbyJoin,
		widget.NewLabel("Create a lobby"),
		lobbyCreate,
	)

	return w
}

func (w *LobbySelectionWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

func joinLobby(lobbyCode string, parentWindow fyne.Window, env env.Env) sharedModels.UserRole {
	joinRequest := sharedModels.LobbyJoinRequest{
		LobbyToken: lobbyCode,
	}
	joinJson, err := json.Marshal(joinRequest)
	if err != nil {
		log.Warn().Msg("failed to marshall login info json")
		dialog.ShowError(err, parentWindow)
		return sharedModels.NoRole
	}

	bodyReader := bytes.NewReader(joinJson)
	req, err := http.NewRequest("POST", env.Url+"/lobby/join", bodyReader)
	if err != nil {
		dialog.ShowError(err, parentWindow)
		return sharedModels.NoRole
	}
	// error handing already handled since it shows a popup
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		log.Warn().Msg("couldn't get app config" + fmt.Sprint(err))
		dialog.ShowError(err, parentWindow)
		return sharedModels.NoRole
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn().Msg("couldn't make request" + fmt.Sprint(err))
		dialog.ShowError(err, parentWindow)
		return sharedModels.NoRole
	}
	switch res.StatusCode {
	case http.StatusOK:
		httpBody, err := helpers.ReadHttpResponse(res.Body)
		if err != nil {
			log.Err(err).Msg("failed to read http response from body")
			dialog.ShowError(err, parentWindow)
			return sharedModels.NoRole
		}
		var joinResponse sharedModels.JoinResponse
		err = json.Unmarshal(httpBody, &joinResponse)
		if err != nil {
			log.Err(err).Msg("failed to unmarshal join response")
			dialog.ShowError(err, parentWindow)
			return sharedModels.NoRole
		}
		appConfig, err := helpers.GetAppConfig(env, parentWindow)
		if err != nil {
			log.Err(err)
			dialog.ShowError(err, parentWindow)
			return sharedModels.NoRole
		}
		appConfig.LobbyToken = lobbyCode
		appConfig.Role = joinResponse.CurrentRole
		// tries to create the user in the db
		result := env.DB.Save(&appConfig)
		if result.Error != nil {
			log.Err(err).Msg("failed to save configuration in database")
			dialog.ShowError(err, parentWindow)
			return sharedModels.NoRole
		}
		log.Info().Msg("joined lobby " + lobbyCode)
		log.Debug().Msg("role is " + fmt.Sprint(joinResponse.CurrentRole))
		switch joinResponse.CurrentRole {
		case sharedModels.NoRole:
			log.Debug().Msg("creating new game widget")
			parentWindow.SetContent(NewGameWidget(env, parentWindow))
		case sharedModels.Hider:
			log.Debug().Msg("creating new hider widget")
			parentWindow.SetContent(NewHiderWidget(env, parentWindow))
		case sharedModels.Seeker:
			log.Debug().Msg("creating new seeker widget")
			parentWindow.SetContent(NewSeekerWidget(env, parentWindow))
		default:
			log.Debug().Msg("unknown role with index " + fmt.Sprint(joinResponse.CurrentRole) + " detected")
		}
	case http.StatusForbidden:
		message := "Not logged in. Log out and back in."
		log.Info().Msg(message)
		error := errors.New(message)
		dialog.ShowError(error, parentWindow)
		return sharedModels.NoRole
	case http.StatusNotFound:
		message := "Lobby doesn't exist"
		log.Info().Msg(message)
		error := errors.New(message)
		dialog.ShowError(error, parentWindow)
		return sharedModels.NoRole
	default:
		dialog.ShowError(errors.New(fmt.Sprint(res.StatusCode)), parentWindow)
		return sharedModels.NoRole
	}
	return sharedModels.NoRole
}
