// +build ignore

package main

import (
	"fmt"
	"github.com/fari-99/fsm"
)

func main() {
	fsmTest := fsm.NewFSM(
		"open",
		fsm.Events{
			{
				Name: "pull",
				Src: []string{"closed"},
				Dst: "open",
				Props:fsm.Properties{
					"editable": true,
					"deletable": false,},},
			{
				Name: "push",
				Src: []string{"closed"},
				Dst: "open",
				Props:fsm.Properties{
					"editable": true,
					"deletable": false,},},
			{
				Name: "close",
				Src: []string{"open"},
				Dst: "closed"},
		},
		fsm.Callbacks{},
	)

	fmt.Println(fsmTest.Current())

	properties := fsmTest.GetPropertiesTransitions()
	fmt.Println(properties)

	//for eventName, propertiesArray := range properties {
	//	fmt.Println("event Name : " + eventName)
	//	fmt.Println("properties : ")
	//
	//	if len(propertiesArray) > 0 {
	//		for propertiesName, propertiesValue := range propertiesArray{
	//			fmt.Println("properties name : " + propertiesName)
	//			fmt.Println("properties value : " + strconv.FormatBool(reflect.ValueOf(propertiesValue).Bool()))
	//		}
	//	} else {
	//		fmt.Println("Event doesn't have properties")
	//	}
	//}
}
