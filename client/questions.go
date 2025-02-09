package client

import (
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/helpers"

	"github.com/jkulzer/fib-server/sharedModels"

	"github.com/jkulzer/osm"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"fyne.io/fyne/v2"
)

func AskRadar(env env.Env, parentWindow fyne.Window, radius float64) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+loginInfo.LobbyToken+"/questions/radar/"+fmt.Sprint(radius), nil)
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
		return errors.New("Lobby doesn't exist")
	case http.StatusForbidden:
		return errors.New("You are not the seeker and can't ask questions")
	default:
		return errors.New("asking radar failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

func AskSameBezirk(env env.Env, parentWindow fyne.Window) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+loginInfo.LobbyToken+"/questions/sameBezirk", nil)
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
		return errors.New("Lobby doesn't exist")
	case http.StatusForbidden:
		return errors.New("You are not the seeker and can't ask questions")
	default:
		return errors.New("asking same bezirk question failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

func AskSameOrtsteil(env env.Env, parentWindow fyne.Window) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+loginInfo.LobbyToken+"/questions/sameOrtsteil", nil)
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
		return errors.New("Lobby doesn't exist")
	case http.StatusForbidden:
		return errors.New("You are not the seeker and can't ask questions")
	default:
		return errors.New("asking same bezirk question failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

func StartThermometer(env env.Env, parentWindow fyne.Window, distance float64) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	thermometerRequest := sharedModels.ThermometerRequest{
		Distance: distance,
	}

	marshalledRequest, err := json.Marshal(thermometerRequest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+loginInfo.LobbyToken+"/questions/thermometer/start", bytes.NewReader(marshalledRequest))
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
		return errors.New("Lobby doesn't exist. Bad Request.")
	case http.StatusForbidden:
		return errors.New("You are not the seeker and can't ask questions")
	case http.StatusConflict:
		return errors.New("You already started a thermometer. Finish the current thermometer first!")
	default:
		return errors.New("asking thermometer question failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

func EndThermometer(env env.Env, parentWindow fyne.Window) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+loginInfo.LobbyToken+"/questions/thermometer/end", nil)
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
		return errors.New("Lobby doesn't exist. Bad Request.")
	case http.StatusForbidden:
		return errors.New("You are not the seeker and can't ask questions")
	case http.StatusMethodNotAllowed:
		return errors.New("You haven't covered the full distance of the thermometer!")
	default:
		return errors.New("asking thermometer question failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

func GetCloseRoutes(env env.Env, parentWindow fyne.Window) (sharedModels.RouteProximityResponse, error) {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return sharedModels.RouteProximityResponse{}, err
	}

	req, err := http.NewRequest("GET", env.Url+"/lobby/"+loginInfo.LobbyToken+"/questions/closeRoutes", nil)
	if err != nil {
		return sharedModels.RouteProximityResponse{}, err
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return sharedModels.RouteProximityResponse{}, err
	}

	switch res.StatusCode {
	case http.StatusOK:

		responseBody, err := helpers.ReadHttpResponse(res.Body)
		if err != nil {
			err := errors.New("couldn't read response body.")
			return sharedModels.RouteProximityResponse{}, err
		}
		var unmarshaledResponse sharedModels.RouteProximityResponse
		err = json.Unmarshal(responseBody, &unmarshaledResponse)
		if err != nil {
			err := errors.New("couldn't unmarshal response body.")
			return sharedModels.RouteProximityResponse{}, err
		}
		return unmarshaledResponse, nil
	case http.StatusBadRequest:
		err := errors.New("Lobby doesn't exist. Bad Request.")
		return sharedModels.RouteProximityResponse{}, err
	case http.StatusForbidden:
		err := errors.New("You are not the seeker and can't ask questions")
		return sharedModels.RouteProximityResponse{}, err
	default:
		err := errors.New("asking same bezirk question failed with http status code " + fmt.Sprint(res.StatusCode))
		return sharedModels.RouteProximityResponse{}, err
	}
}

func AskTrainservice(env env.Env, parentWindow fyne.Window, routeID osm.RelationID) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	trainServiceRequest := sharedModels.TrainServiceRequest{RouteID: routeID}
	requestBytes, err := json.Marshal(trainServiceRequest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+loginInfo.LobbyToken+"/questions/trainService", bytes.NewReader(requestBytes))
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
		err := errors.New("Lobby doesn't exist. Bad Request.")
		return err
	case http.StatusForbidden:
		err := errors.New("You are not the seeker and can't ask questions")
		return err
	default:
		err := errors.New("asking train service question failed with http status code " + fmt.Sprint(res.StatusCode))
		return err
	}
}
