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
		err := row.Scan(&id, &number)

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
				scanErr := row.Scan(&dbCode)

				if dbCode == userCode && scanErr == nil && numErr == nil && codeErr == nil {
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
				_ = row.Scan(&isRequestPresent)

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
				_, err := db.Exec("UPDATE Users SET name=? WHERE id=? AND number=?", name, id, number)

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
		description := r.FormValue("description")

		if isAuthenticated(id, number) {
			if len(location) != 0 && strings.Contains(location, ",") && len(items) != 0 && len(description) != 0 && len(description) <= 140 {
				_, itemErr := strconv.Atoi(items)
				latLng := strings.Split(location, ",")
				lat, latErr := strconv.ParseFloat(latLng[0], 64)
				lng, lngErr := strconv.ParseFloat(latLng[1], 64)

				if itemErr == nil && latErr == nil && lngErr == nil {
					_, err1 := db.Exec("INSERT INTO Transactions VALUES(NOW(), ?, ?, 'open', ?)", number, items, description)
					_, err2 := db.Exec("UPDATE Users SET latitude=?, longitude=? WHERE id=? AND number=?", lat, lng, id, number)

					if err1 == nil && err2 == nil {
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
		displayPage(w, "donate.html")
	}
}

func recentHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id := r.FormValue("id")
		number := r.FormValue("number")
		userLocation := r.FormValue("location")
		radius := r.FormValue("radius")
		status := r.FormValue("status")

		if isAuthenticated(id, number) {
			if len(userLocation) != 0 && strings.Contains(userLocation, ",") && len(radius) != 0 && len(status) != 0 {
				latLng := strings.Split(userLocation, ",")
				lat, latErr := strconv.ParseFloat(latLng[0], 64)
				lng, lngErr := strconv.ParseFloat(latLng[1], 64)
				radius, radErr := strconv.ParseFloat(radius, 64)

				if radErr == nil && latErr == nil && lngErr == nil {
					if strings.Compare(status, "open") == 0 {
						rows, _ := db.Query("SELECT Users.number, name, latitude, longitude, items, description FROM Users JOIN Transactions ON Users.number=Transactions.number WHERE status='open' AND latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?", lat - radius, lat + radius, lng - radius, lng + radius)
						defer rows.Close()

						for rows.Next() {
							var data [6]string
							rows.Scan(&data[0], &data[1], &data[2], &data[3], &data[4], &data[5])
							io.WriteString(w, data[0] + "," + data[1] + "," + data[2] + "," + data[3] + "," + data[4] + "," + data[5] + ";\n")
						}

						io.WriteString(w, approvalMsg)
					} else if strings.Compare(status, "closed") == 0 {
						rows, _ := db.Query("SELECT items, description FROM Users JOIN Transactions ON Users.number=Transactions.number WHERE status='closed' AND latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?", lat - radius, lat + radius, lng - radius, lng + radius)
						defer rows.Close()

						for rows.Next() {
							var data [2]string
							rows.Scan(&data[0], &data[1])
							io.WriteString(w, data[0] + "," + data[1] + ";\n")
						}

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
		displayPage(w, "recent_history.html")
	}
}

func main() {
	var dbErr error
	db, dbErr = sql.Open("mysql", user + ":" + password + "@/" + database)
	defer db.Close()

	if dbErr == nil {
		http.HandleFunc("/login", loginHandler)
		http.HandleFunc("/register", registerHandler)
		http.HandleFunc("/request_sms", requestSmsHandler)
		http.HandleFunc("/set_name", setNameHandler)
		http.HandleFunc("/donate", donateHandler)
		http.HandleFunc("/recent_history", recentHistoryHandler)
		http.ListenAndServe(":8000", nil)
	} else {
		panic(dbErr)
	}
}