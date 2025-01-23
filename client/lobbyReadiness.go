package client

import (
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/helpers"

	"github.com/jkulzer/fib-server/sharedModels"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"fyne.io/fyne/v2"
)

func IsLobbyComplete(env env.Env, parentWindow fyne.Window) (bool, error) {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return false, err
	}
	req, err := http.NewRequest("GET", env.Url+"/lobby/"+loginInfo.LobbyToken+"/readiness", nil)
	if err != nil {
		return false, err
	}
	if err == nil {
		req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return false, err
		} else {
			switch res.StatusCode {
			case http.StatusOK:
				var responseStruct sharedModels.ReadinessResponse
				responseBytes, err := helpers.ReadHttpResponse(res.Body)
				if err != nil {
					return false, err
				}
				err = json.Unmarshal(responseBytes, &responseStruct)
				if err != nil {
					return false, err
				} else {
					return responseStruct.Ready, nil
				}
			default:
				return false, errors.New("readyiness request failed with http status code " + fmt.Sprint(res.StatusCode))
			}
		}
	} else {
		return false, err
	}
}

func SetReadiness(env env.Env, parentWindow fyne.Window, ready bool) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	setReadinessStruct := sharedModels.SetReadinessRequest{
		Ready: ready,
	}
	marshalledJson, err := json.Marshal(setReadinessStruct)
	if err != nil {
		return err
	}
	marshalledJsonReader := bytes.NewReader(marshalledJson)

	req, err := http.NewRequest("PUT", env.Url+"/lobby/"+loginInfo.LobbyToken+"/readiness", marshalledJsonReader)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	switch res.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return errors.New("readiness setting failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}
