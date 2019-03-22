package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	_ "github.com/lib/pq"
)

var db *sql.DB
var tpl *template.Template

type Car struct {
	Name  string
	Brand string
	HP    int //horsepower
	Price float64
}

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://golang:cars@localhost/vehicle?sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Connected to Database")
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	defer db.Close()
	r := httprouter.New()
	r.GET("/cars", getCars) //maybe name
	r.GET("/create", getCreate)
	r.POST("/create", create)
	http.ListenAndServe(":8080", r)
}

func getCars(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	cars, err := db.Query("SELECT cars.NAME, cars.BRAND, cars.HP, cars.PRICE FROM cars")
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	defer cars.Close()

	crows := make([]Car, 0)
	for cars.Next() {
		car := Car{}
		err := cars.Scan(&car.Name, &car.Brand, &car.HP, &car.Price)
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		crows = append(crows, car)
	}

	if err = cars.Err(); err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(w, "cars.html", crows)
}

func getCreate(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tpl.ExecuteTemplate(w, "create_car.html", nil)
}

func create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c := Car{}
	c.Name, c.Brand = r.FormValue("Name"), r.FormValue("Brand")

	if c.Name == "" || c.Brand == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	hp, err := strconv.Atoi(r.FormValue("HP"))
	if err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	price, err := strconv.ParseFloat(r.FormValue("Price"), 64)
	if err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	c.HP, c.Price = hp, price

	_, err = db.Exec("INSERT INTO cars (NAME, BRAND, HP, PRICE) VALUES($1, $2, $3, $4)", c.Name, c.Brand, c.HP, c.Price)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Created!")

}
