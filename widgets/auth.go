package widgets

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/helpers"
	"github.com/jkulzer/fib-client/models"

	"github.com/jkulzer/fib-server/sharedModels"
)

type RegisterWidget struct {
	widget.BaseWidget
	content *fyne.Container
	form    *widget.Form
}

func NewRegisterWidget(env env.Env, parentWindow fyne.Window) *RegisterWidget {
	w := &RegisterWidget{}
	w.ExtendBaseWidget(w)

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("AzureDiamond")
	usernameEntry.Validator = validation.NewRegexp("^.{4,32}$", "Username must be at least 4 or at most 32 characters long")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("hunter2")
	passwordEntry.Validator = validation.NewRegexp("^.{8,32}$", "Password must be at least 8 or at most 32 characters long")

	form := &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{Text: "Username", Widget: usernameEntry},
			{Text: "Password", Widget: passwordEntry},
		},
		OnSubmit: func() { // optional, handle form submission
			go func() {
				registerOnServer(env, usernameEntry.Text, passwordEntry.Text, parentWindow)
				loginOnServer(env, usernameEntry.Text, passwordEntry.Text, parentWindow)
			}()
		},
		SubmitText: "Register",
	}

	w.content = container.NewVBox(
		form,
	)
	return w
}

func (w *RegisterWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

func registerOnServer(env env.Env, username, password string, parentWindow fyne.Window) {
	loginInfo := sharedModels.LoginInfo{
		Username: username,
		Password: password,
	}
	loginInfoJson, err := json.Marshal(loginInfo)
	if err != nil {
		log.Warn().Msg("failed to marshall login info json")
	}

	bodyReader := bytes.NewReader(loginInfoJson)
	res, err := http.Post(env.Url+"/register", "application/json", bodyReader)
	if err != nil {
		log.Warn().Msg("couldn't make request" + fmt.Sprint(err))
		dialog.ShowError(err, parentWindow)
	} else {
		if res.StatusCode == http.StatusCreated {
			log.Info().Msg("user registered")
			dialog.ShowInformation("Registration", "Registration successful!", parentWindow)
		} else if res.StatusCode == http.StatusBadRequest {
			log.Warn().Msg("user already exists")
			error := errors.New("User already exists")
			dialog.ShowError(error, parentWindow)
		} else {
			error := errors.New("failed registering with code " + fmt.Sprint(res.StatusCode))
			dialog.ShowError(error, parentWindow)
		}
	}
}

type LoginWidget struct {
	widget.BaseWidget
	content *fyne.Container
	form    *widget.Form
}

func NewLoginWidget(env env.Env, parentWindow fyne.Window) *LoginWidget {
	w := &LoginWidget{}
	w.ExtendBaseWidget(w)

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("AzureDiamond")
	usernameEntry.Validator = validation.NewRegexp("^.{4,32}$", "Username is at least 4 or at most 32 characters long")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("hunter2")
	passwordEntry.Validator = validation.NewRegexp("^.{8,32}$", "Password is at least 8 or at most 32 characters long")

	form := &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{Text: "Username", Widget: usernameEntry},
			{Text: "Password", Widget: passwordEntry},
		},
		OnSubmit: func() { // optional, handle form submission
			go func() {
				loginOnServer(env, usernameEntry.Text, passwordEntry.Text, parentWindow)
			}()
		},
		SubmitText: "Login",
	}

	w.content = container.NewVBox(
		form,
	)
	return w
}

func (w *LoginWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

func loginOnServer(env env.Env, username, password string, parentWindow fyne.Window) {
	loginInfo := sharedModels.LoginInfo{
		Username: username,
		Password: password,
	}
	loginInfoJson, err := json.Marshal(loginInfo)
	if err != nil {
		log.Warn().Msg("failed to marshall login info json")
	}

	bodyReader := bytes.NewReader(loginInfoJson)
	res, err := http.Post(env.Url+"/login", "application/json", bodyReader)
	if err != nil {
		log.Warn().Msg("couldn't make request" + fmt.Sprint(err))
		dialog.ShowError(err, parentWindow)
	} else {
		switch res.StatusCode {
		case http.StatusOK, http.StatusCreated:
			httpResponse, err := helpers.ReadHttpResponse(res.Body)
			if err != nil {
				log.Warn().Msg("failed to read session token http response: " + fmt.Sprint(err))
			}
			var sessionStruct sharedModels.SessionToken
			err = json.Unmarshal(httpResponse, &sessionStruct)
			if err != nil {
				log.Warn().Msg("failed to unmarshal session token struct on login: " + fmt.Sprint(err))
			}
			log.Info().Msg("user logged in with session token " + fmt.Sprint(sessionStruct.Token) + " which expires at " + fmt.Sprint(sessionStruct.Expiry))

			userName := models.LoginInfo{
				ID:    1,
				Token: sessionStruct.Token,
			}
			// tries to create the user in the db
			result := env.DB.Save(&userName)
			if result.Error != nil {
				log.Warn().Msg("error writing user config to DB with error " + fmt.Sprint(result.Error))
			} else {
				log.Info().Msg("wrote user config to DB")
				// loginInfo, _ := helpers.GetAppConfig(env, parentWindow)
				// fmt.Println(loginInfo)
				parentWindow.SetContent(NewLobbyWidget(env, parentWindow))
			}
		case http.StatusForbidden:
		case http.StatusBadRequest:
			message := "Wrong Password"
			log.Info().Msg(message)
			error := errors.New(message)
			dialog.ShowError(error, parentWindow)
		default:
			dialog.ShowError(errors.New(fmt.Sprint(res.StatusCode)), parentWindow)
		}
	}

}

func GetLoginRegisterTabs(env env.Env, parentWindow fyne.Window) *container.AppTabs {
	register := NewRegisterWidget(env, parentWindow)
	login := NewLoginWidget(env, parentWindow)

	return container.NewAppTabs(
		container.NewTabItem("Register", register),
		container.NewTabItem("Login", login),
	)
}
