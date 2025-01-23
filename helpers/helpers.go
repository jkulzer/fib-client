package helpers

import (
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/models"

	fyne "fyne.io/fyne/v2"

	"github.com/rs/zerolog/log"

	"fmt"
	"io"
)

func ReadHttpResponse(input io.ReadCloser) ([]byte, error) {
	if b, err := io.ReadAll(input); err == nil {
		return b, err
	} else {
		return nil, err
	}
}

func ReadHttpResponseToString(input io.ReadCloser) (string, error) {
	if b, err := io.ReadAll(input); err == nil {
		return string(b), err
	} else {
		return "", err
	}
}

func GetAppConfig(env env.Env, parentWindow fyne.Window) (models.LoginInfo, error) {
	var loginInfo models.LoginInfo
	result := env.DB.First(&loginInfo)
	if result.Error != nil {
		log.Err(result.Error)
		return models.LoginInfo{}, result.Error
	} else if loginInfo.Token.String() == models.NullUuidString {
		log.Warn().Msg("auth token uuid string in app config is null")
		log.Debug().Msg(fmt.Sprint(loginInfo))
		// dialog.ShowError(result.Error, parentWindow)
		return models.LoginInfo{}, result.Error
	} else {
		return loginInfo, nil
	}
}
