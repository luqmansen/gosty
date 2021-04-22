package util

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
)

func DebugStruct(c interface{}) {
	v := reflect.ValueOf(c)
	typeOfS := v.Type()

	log.Debug("Debugging Configuration")
	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("Field: %s\tValue: %v\n", typeOfS.Field(i).Name, v.Field(i).Interface())
	}
}
