package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

type RoleSelectionWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewRoleSelectionWidget(env env.Env, parentWindow fyne.Window, validatedLobbyToken string) *RoleSelectionWidget {
	w := &RoleSelectionWidget{}
	w.ExtendBaseWidget(w)
	w.content = container.NewVBox()
	req, err := http.NewRequest("GET", env.Url+"/lobby/"+validatedLobbyToken+"/roles", nil)
	if err != nil {
		dialog.ShowError(err, parentWindow)
	}
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err == nil {
		req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Warn().Msg("couldn't make request" + fmt.Sprint(err))
			dialog.ShowError(err, parentWindow)
		} else {
			switch res.StatusCode {
			case http.StatusOK:
				var responseStruct sharedModels.RoleAvailability
				responseBytes, err := helpers.ReadHttpResponse(res.Body)
				if err != nil {
					log.Err(err).Msg("failed to read http response for role selection")
					dialog.ShowError(err, parentWindow)
				}
				err = json.Unmarshal(responseBytes, &responseStruct)
				if err != nil {
					message := "Failed to unmarshal http response"
					log.Err(err).Msg(message)
					dialog.ShowError(err, parentWindow)
				}

				w.content = container.NewVBox()
				if len(responseStruct) == 0 {
					w.content.Add(widget.NewLabel("No roles available, lobby full"))
				} else {
					for _, role := range responseStruct {
						var button *widget.Button
						appConfig, err := helpers.GetAppConfig(env, parentWindow)
						if err != nil {
							log.Err(err).Msg("failed to get app config in role selection button for loop")
							dialog.ShowError(err, parentWindow)
							break
						} else {
							if role == sharedModels.Hider {
								button = widget.NewButton("Hider", func() {
									log.Info().Msg("chose hider role")
									err := HandleRoleSelection(env, validatedLobbyToken, parentWindow, appConfig, role)
									if err != nil {
										parentWindow.SetContent(container.NewVBox(NewLobbySelectionWidget(env, parentWindow)))
									} else {
										center := NewHiderWidget(env, parentWindow)
										gameFrame := NewGameFrameWidget(env, parentWindow, center)
										parentWindow.SetContent(gameFrame)
									}
								})
							} else if role == sharedModels.Seeker {
								button = widget.NewButton("Seeker", func() {
									log.Info().Msg("chose seeker role")
									err := HandleRoleSelection(env, validatedLobbyToken, parentWindow, appConfig, role)
									if err != nil {
										parentWindow.SetContent(container.NewVBox(NewLobbySelectionWidget(env, parentWindow)))
									} else {
										center := NewSeekerWidget(env, parentWindow)
										gameFrame := NewGameFrameWidget(env, parentWindow, center)
										parentWindow.SetContent(gameFrame)
									}
								})
							} else {
								message := "Reached unreachable state in role selection"
								log.Err(nil).Msg(message)
								w.content = container.NewVBox(widget.NewLabel(message))
								return w
							}
							w.content.Add(container.NewVBox(button))
						}
					}
				}

			case http.StatusForbidden:
				message := "Not logged in. Log out and back in."
				log.Info().Msg(message)
				error := errors.New(message)
				dialog.ShowError(error, parentWindow)
				w.content = container.NewVBox()
			default:
				dialog.ShowError(errors.New(fmt.Sprint(res.StatusCode)), parentWindow)
			}
		}
	}

	return w
}

func (w *RoleSelectionWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

func HandleRoleSelection(env env.Env, validatedLobbyToken string, parentWindow fyne.Window, appConfig models.LoginInfo, role sharedModels.UserRole) error {
	roleRequestStruct := sharedModels.UserRoleRequest{
		Role: role,
	}

	roleJson, err := json.Marshal(roleRequestStruct)
	roleReader := bytes.NewReader(roleJson)
	if err != nil {
		log.Err(err).Msg("")
		dialog.ShowError(err, parentWindow)
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+validatedLobbyToken+"/selectRole", roleReader)
	if err != nil {
		log.Err(err).Msg(fmt.Sprint(err))
		dialog.ShowError(err, parentWindow)
		return err
	} else {
		req.Header.Add("Authorization", "Bearer "+appConfig.Token.String())
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Err(err).Msg(fmt.Sprint(err))
		dialog.ShowError(err, parentWindow)
	}

	switch res.StatusCode {
	case http.StatusOK:
		appConfig.Role = role
		result := env.DB.Save(&appConfig)
		if result.Error != nil {
			log.Err(result.Error).Msg("failed to save roles in db")
			dialog.ShowError(result.Error, parentWindow)
			return err
		}
		return nil
	case http.StatusConflict:
		error := errors.New("This role has already been selected")
		log.Err(err).Msg("")
		dialog.ShowError(error, parentWindow)
		return err
	default:
		error := errors.New("HTTP request failed with code " + fmt.Sprint(res.StatusCode))
		dialog.ShowError(error, parentWindow)
		return err
	}
}
