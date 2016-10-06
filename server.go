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

func isSmsSender(id string, number string) bool {
	if len(number) != 0 && len(id) != 0 {
		row := db.QueryRow("SELECT id, number from SmsSenders where id=? AND number=?", id, number)
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

				if strings.Compare(dbCode, userCode) == 0 && scanErr == nil && numErr == nil && codeErr == nil {
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
						_, err1 := db.Exec("INSERT INTO Transactions(donationId, timestamp, fromNumber, items, status, description) VALUES(?, ?, ?, ?, 'open', ?)", createDonationCode(timestamp + number),timestamp, number, items, description)
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
						rows, _ := db.Query("SELECT Users.number, name, latitude, longitude, items, description, donationId FROM Users JOIN Transactions ON Users.number=Transactions.fromNumber WHERE status='open' AND toNumber='0' AND latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?", minLat, maxLat, minLng, maxLng)
						defer rows.Close()

						for rows.Next() {
							var data [7]string
							rows.Scan(&data[0], &data[1], &data[2], &data[3], &data[4], &data[5], &data[6])
							io.WriteString(w, data[0] + "," + data[1] + "," + data[2] + "," + data[3] + "," + data[4] + "," + data[5] + "," + data[6] + "\n")
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
		fmt.Println("\nacceptDonationHandler=>\nid:" + id + "\nnumber:" + volunteerNumber + "\ndonationId:" + donationId)

		if isAuthenticated(id, volunteerNumber) {
			if len(donationId) != 0 {
				var donorNumber string
				row := db.QueryRow("SELECT fromNumber FROM Transactions WHERE donationId=? AND status='open' AND toNumber='0'", donationId)
				err := row.Scan(&donorNumber)

				if len(donorNumber) != 0 && err == nil {
					code := createSmsCode("sms")
					_, err1 := db.Exec("INSERT INTO SmsRequest(number, code, type) VALUES(?, ?, 'sms')", volunteerNumber, code)
					_, err2 := db.Exec("UPDATE Transactions SET toNumber=? WHERE donationId=?", volunteerNumber, donationId)

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
		displayWebPage(w, "accept_donation.html")
	}
}

func closeDonationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id := r.FormValue("id")
		number := r.FormValue("number")
		donationId := r.FormValue("donationId")
		code := r.FormValue("code")
		fmt.Println("\ncloseDonationHandler=>\nid:" + id + "\nnumber:" + number + "\ndonationId:" + donationId + "\ncode:" + code)

		if isAuthenticated(id, number) {
			if len(donationId) != 0 && len(code) != 0 {
				_, codeErr := strconv.Atoi(code)
				var dbCode, volunteerNumber string
				row := db.QueryRow("SELECT toNumber FROM Transactions WHERE donationId=?", donationId)
				scanErr := row.Scan(&volunteerNumber)

				if len(volunteerNumber) != 0 && scanErr == nil {
					row := db.QueryRow("SELECT code from SmsRequest WHERE number=? AND isCodeSent='y' AND type='sms' AND code=?", volunteerNumber, code)
					scanErr := row.Scan(&dbCode)

					if len(dbCode) != 0 && scanErr == nil && codeErr == nil {
						_, delSmsErr := db.Exec("DELETE FROM SmsRequest WHERE number=? AND type='sms' AND code=?", volunteerNumber, code)
						_, delUserErr := db.Exec("UPDATE Transactions SET status='closed' WHERE donationId=?", donationId)

						if delSmsErr == nil && delUserErr == nil {
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
		displayWebPage(w, "close_donation.html")
	}
}

func listDonationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		number := r.FormValue("number")
		id := r.FormValue("id")
		donationType := r.FormValue("type")
		fmt.Println("\nlistDonationsHandler=>\nid:" + id + "\nnumber:" + number + "\ntype:" + donationType)

		if isAuthenticated(id, number) {
			if len(donationType) != 0 {
				if strings.Compare(donationType, "donated") == 0 {
					rows, _ := db.Query("SELECT name, items, description, donationId FROM Users JOIN Transactions ON Users.number=Transactions.fromNumber WHERE fromNumber=?", number)
					defer rows.Close()

					for rows.Next() {
						var data [4]string
						rows.Scan(&data[0], &data[1], &data[2], &data[3])
						io.WriteString(w, data[0] + "," + data[1] + "," + data[2] + "," + data[3] + "\n")
					}

					io.WriteString(w, approvalMsg)
				} else if strings.Compare(donationType, "volunteered") == 0 {
					rows, _ := db.Query("SELECT Users.number, name, latitude, longitude, items, description, donationId FROM Users JOIN Transactions ON Users.number=Transactions.fromNumber WHERE toNumber=?", number)
					defer rows.Close()

					for rows.Next() {
						var data [7]string
						rows.Scan(&data[0], &data[1], &data[2], &data[3], &data[4], &data[5], &data[6])
						io.WriteString(w, data[0] + "," + data[1] + "," + data[2] + "," + data[3] + "," + data[4] + "," + data[5] + "," + data[6] + "\n")
					}

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
		displayWebPage(w, "list_donations.html")
	}
}

func cancelDonationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		number := r.FormValue("number")
		id := r.FormValue("id")
		donationId := r.FormValue("donationId")
		fmt.Println("\ncancelDonationHandler=>\nid:" + id + "\nnumber:" + number + "\ndonationId:" + donationId)

		if isAuthenticated(id, number) {
			if len(donationId) != 0 {
				var donorNumber, volunteerNumber string
				row := db.QueryRow("SELECT fromNumber, toNumber FROM Transactions WHERE donationId=? AND status='closed'", donationId)
				err := row.Scan(&donorNumber, &volunteerNumber)

				if strings.Compare(donorNumber, number) == 0 && len(volunteerNumber) != 0 && err == nil {
					_, err1 := db.Exec("INSERT INTO SmsRequest(number, code, type) VALUES(?, ?, 'cancel')", volunteerNumber, donorNumber)
					_, err2 := db.Exec("DELETE FROM Transactions WHERE donationId=?", donationId)

					if err1 == nil && err2 == nil {
						io.WriteString(w, approvalMsg)
					} else {
						io.WriteString(w, errorMsg)
					}
				} else if strings.Compare(volunteerNumber, number) == 0 && len(donorNumber) != 0 && err == nil {
					_, err1 := db.Exec("INSERT INTO SmsRequest(number, code, type) VALUES(?, ?, 'cancel')", donorNumber, volunteerNumber)
					_, err2 := db.Exec("UPDATE Transactions SET toNumber='0' WHERE donationId=?", donationId)

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
		displayWebPage(w, "cancel_donation.html")
	}
}

func pendingSmsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		number := r.FormValue("number")
		id := r.FormValue("id")
		fmt.Println("\npendingSmsHandler=>\nid:" + id + "\nnumber:" + number)

		if isSmsSender(id, number) {
			rows, _ := db.Query("SELECT number, code, type FROM SmsRequest WHERE isCodeSent='n'")
			defer rows.Close()

			for rows.Next() {
				var data [3]string
				rows.Scan(&data[0], &data[1], &data[2])
				io.WriteString(w, data[0] + "," + data[1] + "," + data[2] + "\n")
			}

			io.WriteString(w, approvalMsg)
		} else {
			io.WriteString(w, authFailMsg)
		}
	} else {
		displayWebPage(w, "pending_sms.html")
	}
}

func updateSmsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		number := r.FormValue("number")
		id := r.FormValue("id")
		toNumber := r.FormValue("toNumber")
		code := r.FormValue("code")
		msgType := r.FormValue("type")
		fmt.Println("\nupdateSmsHandler=>\nid:" + id + "\nnumber:" + number + "\ntoNumber:" + toNumber + "\ncode:" + code + "\ntype:" + msgType)

		if isSmsSender(id, number) {
			if len(toNumber) != 0 && len(code) != 0 && len(msgType) != 0 {
				_, err := db.Exec("UPDATE SmsRequest SET isCodeSent='y' WHERE isCodeSent='n' AND number=? AND code=? AND type=?", toNumber, code, msgType)

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
		displayWebPage(w, "update_sms.html")
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
		http.HandleFunc("/close_donation", closeDonationHandler)
		http.HandleFunc("/list_donations", listDonationsHandler)
		http.HandleFunc("/cancel_donation", cancelDonationHandler)
		http.HandleFunc("/pending_sms", pendingSmsHandler)
		http.HandleFunc("/update_sms", updateSmsHandler)
		http.ListenAndServe(":8000", nil)
	} else {
		panic(dbErr)
	}
}
