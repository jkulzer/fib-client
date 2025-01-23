//go:build android
// +build android

package location

import (
	"github.com/paulmach/orb"

	"encoding/json"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver"
)

/*
#include <stdint.h>
#include <stdlib.h>
#include <stdbool.h>

const char *isLocationEnabled(uintptr_t java_vm, uintptr_t jni_env, uintptr_t ctx);
*/
import "C"

type location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func GetLocation(parentWindow fyne.Window) (orb.Point, error) {
	var locationStruct location
	var locationJsonString string
	driver.RunNative(func(ctx interface{}) error {
		ac := ctx.(*driver.AndroidContext)

		str := C.isLocationEnabled(C.uintptr_t(ac.VM), C.uintptr_t(ac.Env), C.uintptr_t(ac.Ctx))
		locationJsonString = C.GoString(str)
		err := json.Unmarshal([]byte(locationJsonString), &locationStruct)
		if err != nil {
			dialog.ShowError(err, parentWindow)
		}

		return nil
	})

	var point orb.Point
	point[0] = locationStruct.Lon
	point[1] = locationStruct.Lat

	return point, nil
}
