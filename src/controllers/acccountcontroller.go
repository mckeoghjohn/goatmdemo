package controllers

import (
	"errors"
	"model"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func FindAccount(db *gorm.DB, r *http.Request) (accountRet model.Account, err error) {
	vars := mux.Vars(r)
	accountNo := vars["account"]
	db.Where("account_number=?", accountNo).Find(&accountRet)

	if accountRet.AccountNumber == 0 {
		err = errors.New("AccountNotFound")
		return accountRet, err
	}
	return accountRet, err
}

func UpdateAccount(db *gorm.DB, account *model.Account) {
	db.Save(&account)
}
