package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
	//"bufio"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)


type context struct {
	repo      *repo
	templates *template.Template
	event     event
	log       *log.Logger
}

type appHandler struct {
	*context	
	Handler func(*context, http.ResponseWriter, *http.Request) (int, error)
	AdminOnly bool
}

func main() {
	//delete the datebase
	if err := os.Remove("./everydaybham.db"); err != nil {
		fmt.Sprintln("ERROR could not remove database:", err)
	}
	//parse templates
	FuncMap := template.FuncMap{"PrettyMonth":PrettyMonth}
	templates := template.Must(template.New("test").Funcs(FuncMap).ParseGlob("templates/*"))

	//create the db
	db, err := sqlx.Connect("sqlite3", "everydaybham.db")
	if err != nil {
		panic(err)
	}
	initDB(db)

	//create logger
	logger := log.New(os.Stdout, "LOG:", log.Lshortfile)

	repo := &repo{db, logger}
	context := &context{repo, templates, event{}, logger}

	//set once at start of program
	setTodaysEvent(context)

	go updateDaily(context)
	//go interactiveConsole(context)

	r := mux.NewRouter()
	r.Handle("/events/createedit", appHandler{context, EventHandler, true})
	r.Handle("/events/createedit/{id}", appHandler{context, EventHandler, true})
	r.Handle("/events", appHandler{context, EventsHandler, true})
	r.Handle("/login", appHandler{context, LoginHandler, false})
	r.Handle("/", appHandler{context, IndexHandler, false})
	fs := http.FileServer(http.Dir("./public/"))
	r.PathPrefix("/").Handler(fs)

	r.NotFoundHandler = appHandler{context, NotFoundHandler, false}
	fmt.Println(http.ListenAndServe(":24601", r)) //+os.Getenv("PORT"), mux))
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ah.AdminOnly{
		cookie, err := r.Cookie("authtoken")	
		if err != nil || cookie.Value != authToken{
			log.Println("error evaluating cookie")
			http.RedirectHandler("/login?url="+r.URL.Path, http.StatusSeeOther).ServeHTTP(w, r)
			return
		}
	}	
	status, err := ah.Handler(ah.context, w, r)
	log.Println("Status:", status, r.URL)	
	if err != nil {
		log.Printf("HTTP %d: %q", status, err)		
		switch status {
		case http.StatusSeeOther:
			return
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
	for { //loop hourly to set between 4 and 5 am( to account for timezones accross the us.)
		now := time.Now()
		fmt.Println(now.Hour())
		if now.Hour() >= 4 && now.Hour() < 5 {
			setTodaysEvent(c)
		}
		time.Sleep(time.Hour)
	}
}
func setTodaysEvent(c *context) {
	e, err := c.repo.getEventByDate(timeToString(time.Now()))
	if err != nil {
		c.log.Println(err)
		c.event = event{-2, "Um, we got nothin'.", time.Now(), "Movie night!", "Helvetica"}
	}
	c.event = e
	return
}

//template functions
func PrettyMonth(m time.Month)string{
	return m.String()[0:3]+"."
}
