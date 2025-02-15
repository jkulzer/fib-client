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

	"github.com/paulmach/orb"

	"fyne.io/fyne/v2"
)

func ValidateAndSetHidingZone(env env.Env, parentWindow fyne.Window, point orb.Point) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	locationRequest := sharedModels.LocationRequest{
		Location: point,
	}
	marshalledJson, err := json.Marshal(locationRequest)
	if err != nil {
		return err
	}
	marshalledJsonReader := bytes.NewReader(marshalledJson)

	req, err := http.NewRequest("PUT", env.Url+"/lobby/"+loginInfo.LobbyToken+"/saveHidingZone", marshalledJsonReader)
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
	case http.StatusBadRequest:
		return errors.New("Invalid Hiding Spot.\nProbably not close enough to a train station (has to be 500 meters) or not in Berlin. Move closer to a train station in Berlin and try again")
	case http.StatusForbidden:
		return errors.New("You are not a hider and therefore cannot set your hiding spot.")
	default:
		return errors.New("setting hiding spot failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

func SaveLocation(env env.Env, parentWindow fyne.Window, point orb.Point) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	locationRequest := sharedModels.LocationRequest{
		Location: point,
	}
	marshalledJson, err := json.Marshal(locationRequest)
	if err != nil {
		return err
	}
	marshalledJsonReader := bytes.NewReader(marshalledJson)

	req, err := http.NewRequest("PUT", env.Url+"/lobby/"+loginInfo.LobbyToken+"/saveLocation", marshalledJsonReader)
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
	case http.StatusBadRequest:
		return errors.New("Can't find lobby")
	case http.StatusForbidden:
		return errors.New("Not authenticated.")
	default:
		return errors.New("saving location failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}
