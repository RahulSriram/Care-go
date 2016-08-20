package main

import (
	"database/sql"
	"html/template"
	"io"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var (
	user string = "test"
	password string = "test"
	database string = "Care"
	db *sql.DB
	err error
	authFailMsg string = "auth_fail"
	errorMsg string = "error"
	approvalMsg string = "ok"
)

func displayPage(w http.ResponseWriter, file string) {
	t, _ := template.ParseFiles(file)
	t.Execute(w, nil)
}

/*func createCode() string {
	//magical code from 'crypt.go'
}*/

func isAuthenticated(id string, number string) bool {
	if len(number) != 0 && len(id) != 0 {
		row := db.QueryRow("SELECT id, number from Users where id=? AND number=?", id, number)
		err = row.Scan(&id, &number)

		if len(id) != 0 && len(number) != 0 && err == nil {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		number := r.FormValue("number")
		id := r.FormValue("id")
		if isAuthenticated(id, number) {
			io.WriteString(w, approvalMsg)
		} else {
			io.WriteString(w, authFailMsg)
		}
	} else {
		displayPage(w, "login.html")
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id := r.FormValue("id")
		number := r.FormValue("number")
		userCode := r.FormValue("code")

		if len(id) != 0 && len(number) != 0 {
			if len(userCode) != 0 {
				//TODO: Add register handling function
				_, errNum := strconv.Atoi(number)
				_, errCode := strconv.Atoi(userCode)
				row := db.QueryRow("SELECT code from SmsRequest WHERE number=? AND isCodeSent='y'", number)
				var dbCode string
				err = row.Scan(&dbCode)

				if dbCode == userCode && err != nil && errNum != nil && errCode != nil {
					db.Exec("DELETE FROM SmsRequest WHERE number=?", number)
					db.Exec("INSERT INTO Users(id, number) VALUES(?, ?)", id, number)
					io.WriteString(w, approvalMsg)
				} else {
					io.WriteString(w, errorMsg)
				}
			} else {
				io.WriteString(w, errorMsg)
			}
		} else {
			io.WriteString(w, authFailMsg)
		}
	} else {
		displayPage(w, "register.html")
	}
}

func requestSmsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		numberString := r.FormValue("number")

		if len(numberString) != 0 {
			number, err := strconv.Atoi(numberString)

			if err != nil {
				var isRequestPresent string

				row := db.QueryRow("SELECT number FROM SmsRequest WHERE number=?", number)
				err = row.Scan(&isRequestPresent)

				if len(isRequestPresent) == 0 {
					code := createCode()
					_, err = db.Exec("INSERT INTO SmsRequest(number, code) VALUES(?, ?)", number, code)

					if err != nil {
						io.WriteString(w, approvalMsg)
					} else {
						io.WriteString(w, errorMsg)
					}
				} else {
					io.WriteString(w, errorMsg)
				}
			} else {
				io.WriteString(w, errorMsg)
			}
		} else {
			io.WriteString(w, errorMsg)
		}
	} else {
		displayPage(w, "request_sms.html")
	}
}

func main() {
	db, err = sql.Open("mysql", user+":"+password+"@/"+database)
	defer db.Close()
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/request_sms", requestSmsHandler)
	http.ListenAndServe(":8000", nil)
}
