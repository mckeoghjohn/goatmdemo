package main

import (
	"controllers"
	"encoding/json"
	"errors"
	"fmt"
	"model"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type App struct {
	Router *mux.Router
	DB     *gorm.DB
}

/*
* Intialize  call to setup db connection and tables
 */
func (a *App) Initialize(dbtype, connectionString string) {

	var err error
	a.DB, err = gorm.Open(dbtype, connectionString)
	if err != nil {
		panic(err)
	}

	a.DB.DropTableIfExists(&model.Account{})
	a.DB.AutoMigrate(&model.Account{})
	a.DB.DropTableIfExists(&model.Notedata{})
	a.DB.CreateTable(&model.Notedata{})
	a.DB.DropTableIfExists(&model.Note{})
	a.DB.CreateTable(&model.Note{})
	a.DB.DropTableIfExists(&model.Atm{})
	a.DB.CreateTable(&model.Atm{})

	a.putCashInAtm()

	a.addUsers()

	a.Router = mux.NewRouter().StrictSlash(true)

	a.InitializeRoutes()
}

func (a *App) addUsers() {
	account1 := model.Account{AccountNumber: 123456, Balance: 600, Overdraft: 600, Pin: 1234}
	account2 := model.Account{AccountNumber: 654321, Balance: 200, Overdraft: 200, Pin: 4321}
	a.DB.Save(&account1)
	a.DB.Save(&account2)
}

func (a *App) putCashInAtm() {

	five := &model.Notedata{Denomination: 5, NumberOfNotes: 10}
	ten := &model.Notedata{Denomination: 10, NumberOfNotes: 10}
	twenty := &model.Notedata{Denomination: 20, NumberOfNotes: 10}
	fifty := &model.Notedata{Denomination: 50, NumberOfNotes: 10}
	a.DB.Save(&five)
	a.DB.Save(&ten)
	a.DB.Save(&twenty)
	a.DB.Save(&fifty)
	notes := make([]model.Note, 4)
	notes[0] = model.Note{Data: five}
	notes[1] = model.Note{Data: ten}
	notes[2] = model.Note{Data: twenty}
	notes[3] = model.Note{Data: fifty}

	atm := model.Atm{Notes: notes}
	a.DB.Save(&atm)
	for i := 0; i < 4; i++ {
		a.DB.Save(&notes[i])
	}
}
func (a *App) Run(port string) error {
	fmt.Printf("\nListening on port %s\n", port)
	return http.ListenAndServe(port, a.Router)
}

func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/account/{account:[0-9]+}", a.listAccountHandler).Methods("GET")
	a.Router.HandleFunc("/account/{account:[0-9]+}/balance", a.balanceHandler).Methods("GET")
	a.Router.HandleFunc("/account/{account:[0-9]+}/limit", a.limitHandler).Methods("GET")
	a.Router.HandleFunc("/account/{account:[0-9]+}/withdraw/{amount:[0-9]+}", a.withdrawHandler).Methods("PUT")
	a.Router.HandleFunc("/account/{account:[0-9]+}/deposit/{amount:[0-9]+}", a.depositHandler).Methods("PUT")
	a.Router.HandleFunc("/account/{account:[0-9]+}/delete", a.deleteHandler).Methods("DELETE")
}

func (a *App) listAccountHandler(w http.ResponseWriter, r *http.Request) {
	account, err := controllers.FindAccount(a.DB, r)

	if err != nil {
		writeHeader(w, http.StatusNotFound)
		encodeErrorJSON(w, fmt.Sprintf("%v", err))
		return
	}

	err = validatePin(w, r, account.Pin)
	if err != nil {
		return
	}
	writeHeader(w, http.StatusOK)
	json.NewEncoder(w).Encode(account)

}

func (a *App) balanceHandler(w http.ResponseWriter, r *http.Request) {
	account, err := controllers.FindAccount(a.DB, r)

	if err != nil {
		writeHeader(w, http.StatusNotFound)
		encodeErrorJSON(w, fmt.Sprintf("%v", err))
		return
	}

	err = validatePin(w, r, account.Pin)
	if err != nil {
		return
	}

	writeHeader(w, http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"Balance": account.Balance})

}

func (a *App) limitHandler(w http.ResponseWriter, r *http.Request) {
	account, err := controllers.FindAccount(a.DB, r)

	if err != nil {
		writeHeader(w, http.StatusNotFound)
		encodeErrorJSON(w, fmt.Sprintf("%v", err))
		return
	}

	err = validatePin(w, r, account.Pin)
	if err != nil {
		return
	}

	writeHeader(w, http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"Limit": account.Balance + account.Overdraft})

}

func (a *App) depositHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Deposit Handler\n")
	account, err := controllers.FindAccount(a.DB, r)

	if err != nil {
		writeHeader(w, http.StatusNotFound)
		encodeErrorJSON(w, fmt.Sprintf("%v", err))
		return
	}

	err = validatePin(w, r, account.Pin)
	if err != nil {
		return
	}

	vars := mux.Vars(r)
	amount := vars["amount"]

	amountVal, err := strconv.ParseInt(amount, 0, 32)
	if err != nil {
		writeHeader(w, http.StatusNotFound)
		encodeErrorJSON(w, fmt.Sprintf("%v", err))
		return
	}
	account.Balance += int(amountVal)

	controllers.UpdateAccount(a.DB, &account)

	writeHeader(w, http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"Balance": account.Balance})
}

func (a *App) withdrawHandler(w http.ResponseWriter, r *http.Request) {
	account, err := controllers.FindAccount(a.DB, r)

	if err != nil {
		writeHeader(w, http.StatusNotFound)
		encodeErrorJSON(w, "Account Not Found")
		return
	}

	err = validatePin(w, r, account.Pin)
	if err != nil {
		return
	}

	vars := mux.Vars(r)
	amount := vars["amount"]
	amountVal, err := strconv.ParseInt(amount, 0, 32)

	if account.Balance+account.Overdraft <= int(amountVal) {
		writeHeader(w, http.StatusInternalServerError)
		encodeErrorJSON(w, "InsufficientFunds")
		return
	}

	notes, err := controllers.Dispense(a.DB, int(amountVal))

	if err != nil {
		writeHeader(w, http.StatusInternalServerError)
		encodeErrorJSON(w, fmt.Sprintf("%v", err))
		return
	}

	writeHeader(w, http.StatusOK)
	json.NewEncoder(w).Encode(notes)

	account.Balance -= int(amountVal)

	controllers.UpdateAccount(a.DB, &account)

}

func (a *App) deleteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Delete Handler\n")
}

func encodeErrorJSON(w http.ResponseWriter, err string) {
	json.NewEncoder(w).Encode(map[string]string{"error": err})
}

func writeHeader(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}

func validatePin(w http.ResponseWriter, r *http.Request, inPin int) error {
	pin := r.FormValue("pin")
	//check pin
	pinVal, _ := strconv.ParseInt(pin, 0, 32)
	if inPin != int(pinVal) {
		err := errors.New("InvalidPin")
		writeHeader(w, http.StatusInternalServerError)
		encodeErrorJSON(w, fmt.Sprintf("%v", err))
		return err
	}
	return nil
}
