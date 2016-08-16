package main

import (
	"io"
	"net/http"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  //"html/template"
)

var arduinoData string

func loginHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    number := r.FormValue("number")
    id := r.FormValue("id")

    if len(number) != 0 && len(id) != 0 {
      database := "Care"
      user := "test"
      password := "test"
      conn, err := sql.Open("mysql", user + ":" + password + "@/" + database)
      defer conn.Close()
      row := conn.QueryRow("SELECT id,number from Users where id=? AND number=?", id, number)
      err = row.Scan(&id, &number)

      if len(id) != 0 && len(number) != 0 && err == nil {
        io.WriteString(w, "ok")
      } else {
        io.WriteString(w, "error")
      }
    }
  }/* else {
    t, _ := template.ParseFiles("login.html")
    t.Execute(w, nil)
  }*/
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.ListenAndServe(":8000", nil)
}
