package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"io"
	"html/template"
	"strconv"
	"strings"
	"fmt"
	"math"
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

func displayWebPage(w http.ResponseWriter, file string) {
	t, _ := template.ParseFiles(file)
	t.Execute(w, nil)
}

/*func createSmsCode(msgType string) string {
	//magical code from 'crypt.go'
}*/

/*func createDonationCode(input string) string {
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
		fmt.Println("\nloginHandler=>\nid:" + id + "\nnumber:" + number)

		if isAuthenticated(id, number) {
			io.WriteString(w, approvalMsg)
		} else {
			io.WriteString(w, authFailMsg)
		}
	} else {
		displayWebPage(w, "login.html")
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id := r.FormValue("id")
		number := r.FormValue("number")
		userCode := r.FormValue("code")
		fmt.Println("\nregisterHandler=>\nid:" + id + "\nnumber:" + number + "\ncode:" + userCode)

		if len(id) != 0 && len(number) != 0 {
			if len(userCode) != 0 {
				_, numErr := strconv.Atoi(number)
				_, codeErr := strconv.Atoi(userCode)
				row := db.QueryRow("SELECT code from SmsRequest WHERE number=? AND isCodeSent='y' AND type='otp'", number)
				var dbCode string
				scanErr := row.Scan(&dbCode)

				if dbCode == userCode && scanErr == nil && numErr == nil && codeErr == nil {
					_, delSmsErr := db.Exec("DELETE FROM SmsRequest WHERE number=? AND type='otp'", number)
					_, delUserErr := db.Exec("DELETE FROM Users WHERE number=?", number)
					_, insErr := db.Exec("INSERT INTO Users(id, number) VALUES(?, ?)", id, number)

					if delSmsErr == nil && delUserErr == nil && insErr == nil {
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
		displayWebPage(w, "register.html")
	}
}

func requestSmsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		number := r.FormValue("number")
		fmt.Println("\nrequestSmsHandler=>\nnumber:" + number)

		if len(number) != 0 && strings.HasPrefix(number, "+") {
			_, err := strconv.Atoi(number)

			if err == nil {
				var isRequestPresent string

				row := db.QueryRow("SELECT number FROM SmsRequest WHERE number=? AND type='otp'", number)
				_ = row.Scan(&isRequestPresent)

				if len(isRequestPresent) == 0 {
					code := createSmsCode("otp")
					_, err = db.Exec("INSERT INTO SmsRequest(number, code, type) VALUES(?, ?, 'otp')", number, code)

					if err == nil {
						io.WriteString(w, approvalMsg)
					} else {
						io.WriteString(w, errorMsg)
					}
				} else {
					_, err = db.Exec("UPDATE SmsRequest SET isCodeSent='n' WHERE number=? AND type='otp'", number)

					if err == nil {
						io.WriteString(w, approvalMsg)
					} else {
						io.WriteString(w, errorMsg)
					}
				}
			} else {
				io.WriteString(w, errorMsg)
			}
		} else {
			io.WriteString(w, errorMsg)
		}
	} else {
		displayWebPage(w, "request_sms.html")
	}
}

func setNameHandler(w http.ResponseWriter, r *http.Request)  {
	if r.Method == "POST" {
		id := r.FormValue("id")
		number := r.FormValue("number")
		name := r.FormValue("name")
		fmt.Println("\nsetNameHandler=>\nid:" + id + "\nnumber:" + number + "\nname:" + name)

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
		displayWebPage(w, "set_name.html")
	}
}

func donateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id := r.FormValue("id")
		number := r.FormValue("number")
		location := r.FormValue("location")
		items := r.FormValue("items")
		description := r.FormValue("description")
		fmt.Println("\ndonateHandler=>\nid:" + id + "\nnumber:" + number + "\nlocation:" + location + "\nitems:" + items + "\ndescription:" + description)

		if isAuthenticated(id, number) {
			if len(location) != 0 && strings.Contains(location, ",") && len(items) != 0 && len(description) != 0 && len(description) <= 140 {
				_, itemErr := strconv.Atoi(items)
				latLng := strings.Split(location, ",")
				lat, latErr := strconv.ParseFloat(latLng[0], 64)
				lng, lngErr := strconv.ParseFloat(latLng[1], 64)

				if itemErr == nil && latErr == nil && lngErr == nil {
					var timestamp string
					row := db.QueryRow("SELECT NOW()")
					err := row.Scan(&timestamp)

					if err == nil {
						_, err1 := db.Exec("INSERT INTO Transactions VALUES(" + createDonationCode(timestamp + number) + ", " + timestamp + ", ?, ?, 'open', ?)", number, items, description)
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
				io.WriteString(w, errorMsg)
			}
		} else {
			io.WriteString(w, authFailMsg)
		}
	} else {
		displayWebPage(w, "donate.html")
	}
}

func recentHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id := r.FormValue("id")
		number := r.FormValue("number")
		userLocation := r.FormValue("location")
		radius := r.FormValue("radius")
		status := r.FormValue("status")
		fmt.Println("\nrecentHistoryHandler=>\nid:" + id + "\nnumber:" + number + "\nlocation:" + userLocation + "\nradius:" + radius + "\nstatus:" + status)

		if isAuthenticated(id, number) {
			if len(userLocation) != 0 && strings.Contains(userLocation, ",") && len(radius) != 0 && len(status) != 0 {
				latLng := strings.Split(userLocation, ",")
				lat, latErr := strconv.ParseFloat(latLng[0], 64)
				lng, lngErr := strconv.ParseFloat(latLng[1], 64)
				radius, radErr := strconv.ParseFloat(radius, 64)

				if radErr == nil && latErr == nil && lngErr == nil {
					minLat := lat - ((1 / 110.6) * radius)
					maxLat := lat + ((1 / 110.6) * radius)
					minLng := lng - ((1 / (111.3 * math.Cos(lat))) * radius)
					maxLng := lng + ((1 / (111.3 * math.Cos(lat))) * radius)

					if strings.Compare(status, "open") == 0 {
						rows, _ := db.Query("SELECT donationId, Users.number, name, latitude, longitude, items, description FROM Users JOIN Transactions ON Users.number=Transactions.number WHERE status='open' AND latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?", minLat, maxLat, minLng, maxLng)
						defer rows.Close()

						for rows.Next() {
							var data [7]string
							rows.Scan(&data[0], &data[1], &data[2], &data[3], &data[4], &data[5], &data[6])
							io.WriteString(w, data[0] + "," + data[1] + "," + data[2] + "," + data[3] + "," + data[4] + "," + data[5] + "," + data[6] + "\n")
						}

						io.WriteString(w, approvalMsg)
					} else if strings.Compare(status, "closed") == 0 {
						rows, _ := db.Query("SELECT items, description FROM Users JOIN Transactions ON Users.number=Transactions.number WHERE status='closed' AND latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?", lat - radius, lat + radius, lng - radius, lng + radius)
						defer rows.Close()

						for rows.Next() {
							var data [2]string
							rows.Scan(&data[0], &data[1])
							io.WriteString(w, data[0] + "," + data[1] + "\n")
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
		displayWebPage(w, "recent_history.html")
	}
}

func acceptDonationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		volunteerNumber := r.FormValue("number")
		id := r.FormValue("id")
		donationId := r.FormValue("donationId")

		if isAuthenticated(id, volunteerNumber) {
			var donorNumber string
			row := db.QueryRow("SELECT number FROM Transactions WHERE donationId=?", donationId)
			err := row.Scan(&donorNumber)

			if err == nil {
				code := createSmsCode("sms")
				_, err1 := db.Exec("INSERT INTO SmsRequest(number, code, type) VALUES(?, ?, 'sms')", volunteerNumber, code)
				_, err2 := db.Exec("INSERT INTO SmsRequest(number, code, type) VALUES(?, ?, 'sms')", donorNumber, code)
				_, err3 := db.Exec("UPDATE Transactions SET status='closed' WHERE donationId=?", donationId)

				if err1 == nil && err2 == nil && err3 == nil {
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
		displayWebPage(w, "accept_donation.html")
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
		http.HandleFunc("/accept_donation", acceptDonationHandler)
		http.ListenAndServe(":8000", nil)
	} else {
		panic(dbErr)
	}
}
