package main

import (
	"log"
	"net/http"
	"html/template"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"strconv"
)

var db *sql.DB

var (
	key = []byte("secret-key")
	store = sessions.NewCookieStore(key)
)

func home(writer http.ResponseWriter, request *http.Request){
	session, err := store.Get(request, "cookie-name")
	if err != nil{
		log.Fatal(err)
	}
	auth := session.Values["authenticated"].(bool)
	if auth {
		html := template.Must(template.ParseFiles("templates/base.html", "templates/home.html"))
		html.Execute(writer, auth)
	} else {
		html := template.Must(template.ParseFiles("templates/base.html", "templates/home.html"))
		html.Execute(writer, nil)
	}
}

func create(writer http.ResponseWriter, request *http.Request){
	switch request.Method{
		case "POST":
			tx, err := db.Begin()
			if err != nil{
				log.Fatal(err)
			}
			cmd := "INSERT INTO Users"+"(name, username, password)"+"VALUES"+"(?, ?, ?)"

			stmt, err := tx.Prepare(cmd)
			if err != nil{
				log.Fatal(err)
			}
			defer stmt.Close()
	
			if len(request.FormValue("name")) == 0 || len(request.FormValue("username")) == 0 || len(request.FormValue("password")) == 0 {
				html := template.Must(template.ParseFiles("templates/base.html", "templates/create.html"))
				html.Execute(writer, nil)
			} else {
				newUsername := request.FormValue("username")

				stmt.Exec(request.FormValue("name"), newUsername, request.FormValue("password"))
				tx.Commit()

				session, err := store.Get(request, "cookie-name")
				if err != nil{
					log.Fatal(err)
				}
				session.Values["authenticated"] = true

				cmd = "SELECT * FROM Users WHERE username = ?"

				rows, err := db.Query(cmd, newUsername)
				if err != nil{
					log.Fatal(err)
				}
				defer rows.Close()

				for rows.Next(){
					var id int
					var name string
					var username string
					var password string
					err = rows.Scan(&id, &name, &username, &password)
					if err != nil{
						log.Fatal(err)
					}
					session.Values["id"] = id
				}
				session.Save(request, writer)

				auth := true
				html := template.Must(template.ParseFiles("templates/base.html", "templates/home.html"))
				html.Execute(writer, auth)
			}
		default:
			html := template.Must(template.ParseFiles("templates/base.html", "templates/create.html"))
			html.Execute(writer, nil)
	}
}

func login(writer http.ResponseWriter, request *http.Request){
	switch request.Method{
		case "POST":
			session, err := store.Get(request, "cookie-name")
			if err != nil{
				log.Fatal(err)
			}
			var userPassword string
			cmd := "SELECT * FROM Users WHERE username = ?"
			currentUsername := request.FormValue("username")
			if len(currentUsername) == 0{
				html := template.Must(template.ParseFiles("templates/base.html", "templates/login.html"))
				html.Execute(writer, nil)
			} else {
				rows, err := db.Query(cmd, currentUsername)
				if err != nil{
					log.Fatal(err)
				}
				defer rows.Close()

				for rows.Next(){
					var id int
					var name string
					var username string
					var password string
					err = rows.Scan(&id, &name, &username, &password)
					if err != nil{
						log.Fatal(err)
					}
					session.Values["id"] = id
					userPassword = password
				}

				currentPassword := request.FormValue("password")
				if len(currentPassword) == 0 {
					html := template.Must(template.ParseFiles("templates/base.html", "templates/login.html"))
					html.Execute(writer, nil)
				} else if(userPassword == currentPassword){
					session.Values["authenticated"] = true
					session.Save(request, writer)

					auth := true
					html := template.Must(template.ParseFiles("templates/base.html", "templates/home.html"))
					html.Execute(writer, auth)
				} else {
					html := template.Must(template.ParseFiles("templates/base.html", "templates/login.html"))
					html.Execute(writer, nil)
				}
			}
		default:
			html := template.Must(template.ParseFiles("templates/base.html", "templates/login.html"))
			html.Execute(writer, nil)
	}
}

func logout(writer http.ResponseWriter, request *http.Request){
	session, err := store.Get(request, "cookie-name")
				if err != nil{
					log.Fatal(err)
				}
				session.Values["authenticated"] = false
				session.Save(request, writer)

	html := template.Must(template.ParseFiles("templates/base.html", "templates/home.html"))
	html.Execute(writer, nil)
}

type Car struct{
	Id int
	Make string
	Model string
	Year int
	Color string
}

func cars(writer http.ResponseWriter, request *http.Request){
	session, err := store.Get(request, "cookie-name")
	if err != nil{
		log.Fatal(err)
	}
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Error(writer, "Forbidden", http.StatusForbidden)
		return
	}

	cmd := "SELECT * FROM Cars"
	rows, err := db.Query(cmd)

	carList := make([]Car,0)

	for rows.Next(){
		var id int
		var make string
		var model string
		var year int
		var color string
		var user_id int
		err = rows.Scan(&id, &make, &model, &year, &color, &user_id)
		if err != nil{
			log.Fatal(err)
		}
		var car Car
		car.Id = id
		car.Make = make
		car.Model = model
		car.Year = year
		car.Color = color
		if session.Values["id"] == user_id{
			carList = append(carList, car)
		}
	}

	if len(carList) == 0 {
		html := template.Must(template.ParseFiles("templates/base.html", "templates/cars.html"))
		html.Execute(writer, 1)
	} else {
		html := template.Must(template.ParseFiles("templates/base.html", "templates/cars.html"))
		html.Execute(writer, carList)
	}
}

func createCar(writer http.ResponseWriter, request *http.Request){
	session, err := store.Get(request, "cookie-name")
	if err != nil{
		log.Fatal(err)
	}
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Error(writer, "Forbidden", http.StatusForbidden)
		return
	}

	switch request.Method{
		case "POST":
			tx, err := db.Begin()
			if err != nil{
				log.Fatal(err)
			}
			cmd := "INSERT INTO Cars"+"(make, model, year, color, user_id)"+"VALUES"+"(?, ?, ?, ?, ?)"

			stmt, err := tx.Prepare(cmd)
			if err != nil{
				log.Fatal(err)
			}
			defer stmt.Close()
	
			if len(request.FormValue("make")) == 0 || len(request.FormValue("model")) == 0 || len(request.FormValue("year")) == 0 || len(request.FormValue("color")) == 0{
				html := template.Must(template.ParseFiles("templates/base.html", "templates/createCars.html"))
				html.Execute(writer, auth)
			} else {
				stmt.Exec(request.FormValue("make"), request.FormValue("model"), request.FormValue("year"), request.FormValue("color"), session.Values["id"])
				tx.Commit()

				http.Redirect(writer, request, "/cars", http.StatusSeeOther)
			}
		default:
			html := template.Must(template.ParseFiles("templates/base.html", "templates/createCars.html"))
			html.Execute(writer, auth)
	}
}

func delete(writer http.ResponseWriter, request *http.Request){
	session, err := store.Get(request, "cookie-name")
	if err != nil{
		log.Fatal(err)
	}
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Error(writer, "Forbidden", http.StatusForbidden)
		return
	}

	params := request.URL.Query()["id"]
	id := params[0]

	cmd := "DELETE FROM Cars WHERE id = ?"
	db.Exec(cmd, id)

	http.Redirect(writer, request, "/cars", http.StatusSeeOther)
}

func update(writer http.ResponseWriter, request *http.Request){
	session, err := store.Get(request, "cookie-name")
	if err != nil{
		log.Fatal(err)
	}
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Error(writer, "Forbidden", http.StatusForbidden)
		return
	}

	switch request.Method{
	case "POST":
		paramId := request.URL.Query()["id"]
		id,_ := strconv.Atoi(paramId[0])

		make := request.FormValue("make")
		model := request.FormValue("model")
		year := request.FormValue("year")
		color := request.FormValue("color")

		cmd := "UPDATE Cars SET make = ? WHERE id = ?"
		db.Exec(cmd, make, id)

		cmd = "UPDATE Cars SET model = ? WHERE id = ?"
		db.Exec(cmd, model, id)

		cmd = "UPDATE Cars SET year = ? WHERE id = ?"
		db.Exec(cmd, year, id)

		cmd = "UPDATE Cars SET color = ? WHERE id = ?"
		db.Exec(cmd, color, id)

		http.Redirect(writer, request, "/cars", http.StatusSeeOther)
	default:
		var car Car

		paramId := request.URL.Query()["id"]
		car.Id,_ = strconv.Atoi(paramId[0])
		paramMake := request.URL.Query()["make"]
		car.Make = paramMake[0]
		paramModel := request.URL.Query()["model"]
		car.Model = paramModel[0]
		paramYear := request.URL.Query()["year"]
		car.Year,_ = strconv.Atoi(paramYear[0])
		paramColor := request.URL.Query()["color"]
		car.Color = paramColor[0]
		
		html := template.Must(template.ParseFiles("templates/base.html", "templates/update.html"))
		html.Execute(writer, car)
	}
}

func main() {
	db = connect()
	defer db.Close()
	http.HandleFunc("/", home)
	http.HandleFunc("/create", create)
	http.HandleFunc("/login", login)
	http.HandleFunc("/cars", cars)
	http.HandleFunc("/createCar", createCar)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/delete", delete)
	http.HandleFunc("/update", update)
	err := http.ListenAndServe("localhost:5000", nil)
	log.Fatal(err)
}

func connect() *sql.DB {
	db, err := sql.Open("sqlite3", "./final.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	return db
}
