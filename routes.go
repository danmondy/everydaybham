package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"github.com/gorilla/mux"
)
var authToken string

func LoginHandler(c *context, w http.ResponseWriter, r *http.Request) (int, error){
	switch r.Method {
	case "GET":
		url := ""
		url = r.URL.Query()["url"][0]
		
		return c.renderTemplate(w, "login", url)
	case "POST":
		if err := r.ParseForm(); err != nil {
			return c.renderTemplate(w, "login", err.Error())
		}
		email := r.FormValue("email")
		pword := r.FormValue("password")
		url := r.FormValue("url")
		//TODO: encrypt password
		if email == "danmondy@gmail.com" && pword == "unsecure" {
			cookie := &http.Cookie{}
			cookie.Name = "authtoken"
			authToken = time.Now().Format(time.RFC822) //TODO: add random characters to end
			cookie.Value = authToken // needs mutex
			cookie.Expires = time.Now().Add(time.Hour)
			http.SetCookie(w, cookie)
			if url != ""{
				http.Redirect(w, r, url, http.StatusSeeOther)
			}
			http.Redirect(w, r, "/events", http.StatusSeeOther)
			return http.StatusSeeOther, nil
		}else{
			return c.renderTemplate(w, "login", "Username or password not found.")
		}		
	}
	return http.StatusMethodNotAllowed, nil
}
func IndexHandler(c *context, w http.ResponseWriter, r *http.Request) (int, error) {
	fmt.Println("Index Handler Reached")
	fmt.Println(c.event)
	return c.renderTemplate(w, "index", c.event)
}
func EventHandler(c *context, w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.Method {
	case "GET":
		c.log.Println("get reached")
		vars := mux.Vars(r)
		idstring := vars["id"]
		//Create a new event without an id
		fmt.Println(idstring)					
		
		id, err := strconv.Atoi(idstring)
		if err != nil {//if the id does not convert assume they are creating a new entry
			return c.renderTemplate(w, "createedit", event{Date: time.Now()})		
		}
		//retrieve existing event for update
		event, err := c.repo.getEvent(id)
		if err != nil {
			fmt.Println(id, err)
			return http.StatusInternalServerError, errors.New("We could not find an event with the id specified.")
		}
		return c.renderTemplate(w, "createedit", event)
	case "POST":
		c.log.Println("post reached")
		r.ParseForm()
		event := event{}
		event.Description = r.FormValue("description")
		event.Font = r.FormValue("font")
		idstring := r.FormValue("id")

		event.Date, _ = stringToTime(r.FormValue("date"))
		if idstring == "0" {
			fmt.Println("About to insert:", event)
			count, err := c.repo.insertEvent(&event)
			if err != nil || count == 0 {
				c.log.Println(err)
				
			}
		} else {
			event.ID, _ = strconv.Atoi(idstring)
			count, err := c.repo.updateEvent(&event)
			fmt.Println("About to update:", event)
			if err != nil || count == 0 {
				c.log.Println(count, err)				
			}
		}
		return c.renderTemplate(w, "createedit", event)				
	default:
		return http.StatusMethodNotAllowed, errors.New("Bad Method")
	}
}
func EventsHandler(c *context, w http.ResponseWriter, r *http.Request) (int, error) {
	events, err := c.repo.getEventsAfter(time.Now())
	if err != nil {
		return http.StatusInternalServerError, errors.New("Error retrieving events")
	}
	return c.renderTemplate(w, "events", events)
}
func NotFoundHandler(c *context, w http.ResponseWriter, r *http.Request) (int, error) {
	return http.StatusNotFound, errors.New("Not found.")
}
