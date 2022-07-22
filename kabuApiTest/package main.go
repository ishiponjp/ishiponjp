package main

import (
	"fmt"
	"local.packeges/com"
)

func main() {

	wToken, err := com.GetToken(true, "kumi0312")
	if err != nil {
		panic("Error")
	}
	fmt.Println(wToken)

	wPosion, err2 := com.GetPositions(true,wToken)
	if err2 != nil {
		panic("Error2")
	}
	fmt.Println(wPosion)

	wGetRegulations, err := com.GetRegulations(true,wToken, "1694")
	if err != nil {
		panic("Error3")
	}
	fmt.Println(wGetRegulations)

	wGetOrders, err := com.GetOrders(true,wToken)
	if err != nil {
		panic("Error4")
	}
	fmt.Println(wGetOrders)

	wGetMargin, err := com.GetMargin(true,wToken)
	if err != nil {
		panic("Error5")
	}
	fmt.Println(wGetMargin)

	wGetCash, err := com.GetCash(true,wToken)
	if err != nil {
		panic("Error6")
	}
	fmt.Println(wGetCash)

	wGetMarginpremium, err := com.GetMarginpremium(true,wToken, "1694")
	if err != nil {
		panic("Error6")
	}
	fmt.Println(wGetMarginpremium)

}
