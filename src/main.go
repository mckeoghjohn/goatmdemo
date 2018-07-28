package main

import (
	"fmt"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	fmt.Println("Lets Start our Atm")

	a := App{}

	a.Initialize("mysql", "johnmck:password@tcp(localhost:3306)/atmdemo?charset=utf8&parseTime=True&loc=Local")

	a.Run(":8089")

}
