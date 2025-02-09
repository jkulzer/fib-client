package client

import (
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/helpers"

	"net/http"

	"fyne.io/fyne/v2"
)

func GetMapData(env env.Env, parentWindow fyne.Window) ([]byte, error) {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", env.Url+"/lobby/"+loginInfo.LobbyToken+"/map", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := helpers.ReadHttpResponse(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
