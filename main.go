package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
	"bufio"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type event struct {
	ID          int
	Date        time.Time
	Description string
	Font        string
}
type context struct {
	repo      *repo
	templates *template.Template
	event     event
}
type appHandler struct {
	*context
	Ha func(*context, http.ResponseWriter, *http.Request) (int, error)
}

func main() {	
	templates := template.Must(template.ParseGlob("templates/*"))
	
	db, err := sqlx.Connect("sqlite3", "everydaybham.db")
	if err != nil {
		panic(err)
	}
	//os.Remove("everydaybham.db")
	//initDB(db)
	repo := &repo{db}
	context := &context{repo, templates, event{}}

	go updateDaily(context)
	go interactiveConsole(context)

	r := mux.NewRouter()
	r.Handle("/", appHandler{context, IndexHand})
	r.Handle("/createedit", appHandler{context, NewHand})
	r.Handle("/createedit/{id}", appHandler{context, NewHand})
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	r.NotFoundHandler = appHandler{context, NotFoundHand}
	fmt.Println(http.ListenAndServe(":24601", r)) //+os.Getenv("PORT"), mux))
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := ah.Ha(ah.context, w, r)
	fmt.Println("Status:", status, r.URL)
	if err != nil {
		fmt.Printf("HTTP %d: %q", status, err)
		switch status {
		case http.StatusNotFound:
			if _, err := ah.context.renderTemplate(w, "404", nil); err != nil {
				fmt.Println("Error rendering the not found page: ", err)
				http.NotFound(w, r)
			}
		case http.StatusInternalServerError:
			http.Error(w, err.Error(), status)
		default:
			http.Error(w, http.StatusText(status), status)
		}
	}
}
func (c *context) renderTemplate(w http.ResponseWriter, tmpl string, model interface{}) (int, error) {
	err := c.templates.ExecuteTemplate(w, tmpl+".html", model)
	if err != nil {
		fmt.Println(err)
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, err
}

//polls hourly updates between 4 and 5 oclock
func updateDaily(c *context) {
	//set once at start of program
	if err := dailyEvent(c); err != nil {
		panic(err) // panic on startup
	}
	for { //loop hourly to set between 4 and 5 am( to account for timezones accross the us.)
		now := time.Now()
		fmt.Println(now.Hour())
		if now.Hour() >= 4 && now.Hour() < 5 {
			err := dailyEvent(c)
			if err != nil {
				fmt.Println(err)
			}
		}
		time.Sleep(time.Hour)
	}
}
func dailyEvent(c *context) (err error) {
	c.event, err = c.repo.getEventByDate(timeToString(time.Now()))
	return
}

var schema = `
CREATE TABLE event (    
    id   integer unique not null primary key,
    date text unique,
    description text,
    font text
);
`

func initDB(db *sqlx.DB) {
	db.MustExec(schema)
	tx := db.MustBegin()
	tx.MustExec("INSERT into event (date, description, font) values ($1, $2, $3)", timeToString(time.Now()), "Do something fun!", "Helvetica")
	tx.MustExec("INSERT into event (date, description, font) values ($1, $2, $3)", timeToString(time.Now().Add(time.Hour*24)), "Do something Else!", "Helvetica")
	tx.Commit()
}

func interactiveConsole(c *context) {
	
	bio := bufio.NewReader(os.Stdin)
	for {
		cmd, _, _ := bio.ReadLine()
		q, e := c.repo.db.Query(string(cmd))
		if e != nil {
			fmt.Println(e)
		} else {
			for q.Next(){				
				fmt.Println(q.Columns)
			}
			
		}
	}
}
