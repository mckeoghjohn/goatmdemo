package main_test

import (
	"encoding/json"
	"fmt"
	"model"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"."
)

var a main.App
var testAccount model.Account

func TestMain(m *testing.M) {

	testAccount = model.Account{
		AccountNumber: 123456,
		Pin:           1234,
		Balance:       600,
		Overdraft:     600,
	}

	a = main.App{}

	a.Initialize("mysql", "johnmck:password@tcp(localhost:3306)/atmdemo?charset=utf8&parseTime=True&loc=Local")

	code := m.Run()

	os.Exit(code)

}

func reInitialise() {
	a.DB.Close()
	a.Initialize("mysql", "johnmck:password@tcp(localhost:3306)/atmdemo?charset=utf8&parseTime=True&loc=Local")
}

func TestAccountExists(t *testing.T) {
	account := "123456"
	var uriBuilder strings.Builder
	uriBuilder.WriteString("/account/")
	uriBuilder.WriteString(account)
	uriBuilder.WriteString("?pin=1234")
	req, _ := http.NewRequest("GET", uriBuilder.String(), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	accountNum, _ := strconv.Atoi(account)

	accountRet := int(m["AccountNumber"].(float64))
	if accountRet != accountNum {
		t.Errorf("Expected the acccount to be %v but got %v", accountNum, m["AccountNumber"])
	}
}

func TestAccountData(t *testing.T) {
	var uriBuilder strings.Builder
	uriBuilder.WriteString("/account/")
	accountNum := strconv.Itoa(testAccount.AccountNumber)
	uriBuilder.WriteString(accountNum)
	uriBuilder.WriteString("?pin=1234")
	req, _ := http.NewRequest("GET", uriBuilder.String(), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	accountRet := int(m["AccountNumber"].(float64))
	if accountRet != testAccount.AccountNumber {
		t.Errorf("Expected the acccount to be %v but got %v", testAccount.AccountNumber, m["AccountNumber"])
	}

	pinRet := int(m["Pin"].(float64))
	if pinRet != testAccount.Pin {
		t.Errorf("Expected the acccount to be %v but got %v", testAccount.Pin, m["Pin"])
	}

	balanceRet := int(m["Balance"].(float64))
	if balanceRet != testAccount.Balance {
		t.Errorf("Expected the acccount to be %v but got %v", testAccount.Balance, m["Balance"])
	}

	overdraftRet := int(m["Overdraft"].(float64))
	if overdraftRet != testAccount.Overdraft {
		t.Errorf("Expected the acccount to be %v but got %v", testAccount.Overdraft, m["Overdraft"])
	}
}

func TestGetBalance(t *testing.T) {
	var uriBuilder strings.Builder
	uriBuilder.WriteString("/account/")
	accountNum := strconv.Itoa(testAccount.AccountNumber)
	uriBuilder.WriteString(accountNum)
	uriBuilder.WriteString("/balance?pin=1234")
	req, _ := http.NewRequest("GET", uriBuilder.String(), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	balanceRet := int(m["Balance"].(float64))
	if balanceRet != testAccount.Balance {
		t.Errorf("Expected the acccount Balance to be %v but got %v", testAccount.Balance, m["Balance"])
	}
}

func TestGetMaxWithDrawalLimit(t *testing.T) {
	var uriBuilder strings.Builder
	uriBuilder.WriteString("/account/")
	accountNum := strconv.Itoa(testAccount.AccountNumber)
	uriBuilder.WriteString(accountNum)
	uriBuilder.WriteString("/limit?pin=1234")

	req, _ := http.NewRequest("GET", uriBuilder.String(), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	limitRet := int(m["Limit"].(float64))
	actualLimit := testAccount.Overdraft + testAccount.Balance
	if actualLimit != limitRet {
		t.Errorf("Expected the acccount limit to be %v but got %v", actualLimit, limitRet)
	}

}

func TestWithDrawalAndValidateNotes(t *testing.T) {

	var uriBuilder strings.Builder
	uriBuilder.WriteString("/account/")
	accountNum := strconv.Itoa(testAccount.AccountNumber)
	uriBuilder.WriteString(accountNum)
	uriBuilder.WriteString("/withdraw/285?pin=1234") // 285 = (5 * 50) + (1*20) + (1*10) + (1*5)
	fmt.Println(uriBuilder.String())
	testAccount.Balance -= 285
	req, _ := http.NewRequest("PUT", uriBuilder.String(), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var notes []model.Note
	json.Unmarshal(response.Body.Bytes(), &notes)

	for i := range notes {
		switch notes[i].Data.Denomination {
		case 5:
			expectedRes := 1
			checkNotesReturned(t, expectedRes, notes[i])
		case 10:
			expectedRes := 1
			checkNotesReturned(t, expectedRes, notes[i])
		case 20:
			expectedRes := 1
			checkNotesReturned(t, expectedRes, notes[i])
		case 50:
			expectedRes := 5
			checkNotesReturned(t, expectedRes, notes[i])

		default:
			t.Errorf("Invalid Note Returned .. %v", notes[i].Data.Denomination)
		}
	}
}

func TestInsufficientFundsInAccount(t *testing.T) {
	req1 := "/account/654321"
	req1 += "/withdraw/800?pin=4321"

	req, _ := http.NewRequest("PUT", req1, nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusInternalServerError, response.Code)
	checkErrorStringinResponse(t, "InsufficientFunds", response)

}

func TestInsufficientNotesInAtm(t *testing.T) {
	reInitialise()
	req1 := "/account/"
	accountNum := strconv.Itoa(testAccount.AccountNumber)
	req1 += accountNum
	req1 += "/withdraw/800?pin=1234"
	req2 := "/account/"
	req2 += accountNum
	req2 += "/withdraw/100?pin=1234"

	req, _ := http.NewRequest("PUT", req1, nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("PUT", req2, nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusInternalServerError, response.Code)
	checkErrorStringinResponse(t, "InsufficientNotesInAtm", response)
}

func TestInvalidPin(t *testing.T) {
	account := "123456"
	var uriBuilder strings.Builder
	uriBuilder.WriteString("/account/")
	uriBuilder.WriteString(account)
	uriBuilder.WriteString("?pin=12345")
	req, _ := http.NewRequest("GET", uriBuilder.String(), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, response.Code)
	checkErrorStringinResponse(t, "InvalidPin", response)

}

func TestAccountDoesNotExist(t *testing.T) {
	var uriBuilder strings.Builder
	uriBuilder.WriteString("/account/99999")
	req, _ := http.NewRequest("GET", uriBuilder.String(), nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func checkNotesReturned(t *testing.T, expectedRes int, note model.Note) {
	if note.Data.NumberOfNotes != expectedRes {
		t.Errorf("Expected there to be %v but got %v of %v", expectedRes, note.Data.NumberOfNotes, note.Data.Denomination)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func checkErrorStringinResponse(t *testing.T, expected string, response *httptest.ResponseRecorder) {
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	errorString := m["error"].(string)
	if strings.Compare(expected, errorString) != 0 {
		t.Errorf("Expected response string %s. Got %s\n", expected, errorString)
	}
}
