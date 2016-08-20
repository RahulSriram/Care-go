package main

import (
	"database/sql"
	"html/template"
	"io"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"strings"
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

/*func createSmsCode() string {
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
				_, numErr := strconv.Atoi(number)
				_, codeErr := strconv.Atoi(userCode)
				row := db.QueryRow("SELECT code from SmsRequest WHERE number=? AND isCodeSent='y'", number)
				var dbCode string
				err = row.Scan(&dbCode)

				if dbCode == userCode && err == nil && numErr == nil && codeErr == nil {
					_, delErr := db.Exec("DELETE FROM SmsRequest WHERE number=?", number)
					_, insErr := db.Exec("INSERT INTO Users(id, number) VALUES(?, ?)", id, number)

					if delErr == nil && insErr == nil {
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

			if err == nil {
				var isRequestPresent string

				row := db.QueryRow("SELECT number FROM SmsRequest WHERE number=?", number)
				err = row.Scan(&isRequestPresent)

				if len(isRequestPresent) == 0 {
					code := createSmsCode()
					_, err = db.Exec("INSERT INTO SmsRequest(number, code) VALUES(?, ?)", number, code)

					if err == nil {
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

func setNameHandler(w http.ResponseWriter, r *http.Request)  {
	if r.Method == "POST" {
		id := r.FormValue("id")
		number := r.FormValue("number")
		name := r.FormValue("name")

		if isAuthenticated(id, number) {
			if len(name) != 0 {
				_, err = db.Exec("UPDATE Users SET name=? WHERE id=? AND number=?", name, id, number)

				if err == nil {
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
		displayPage(w, "set_name.html")
	}
}

func donateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id := r.FormValue("id")
		number := r.FormValue("number")
		location := r.FormValue("location")
		items := r.FormValue("items")

		if isAuthenticated(id, number) {
			if len(location) != 0 && strings.Contains(location, ",") && len(items) != 0 {
				_, err1 := db.Exec("INSERT INTO Transactions VALUES(?, ?)", number, items)
				_, err2 := db.Exec("UPDATE Users SET location=? WHERE id=? AND number=?", location, id, number)

				if err1 == nil && err2 == nil {
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
		displayPage(w, "donate.html")
	}
}

func main() {
	db, err = sql.Open("mysql", user+":"+password+"@/"+database)
	defer db.Close()
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/request_sms", requestSmsHandler)
	http.HandleFunc("/set_name", setNameHandler)
	http.HandleFunc("/donate", donateHandler)
	http.ListenAndServe(":8000", nil)
}
