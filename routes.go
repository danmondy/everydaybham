package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

//"/"
func IndexHand(c *context, w http.ResponseWriter, r *http.Request) (int, error) {
	fmt.Println("Index Handler Reached")
	fmt.Println(c.event)
	return c.renderTemplate(w, "index", c.event)
}
//"/createedit/{id}" id=new will create a new event object
func NewEventHand(c *context, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == "GET" {
		vars := mux.Vars(r)
		idstring := vars["id"]
		//Create a new event without an id
		fmt.Println(idstring)
		if idstring == "new"{
			return c.renderTemplate(w, "new", event{Date:time.Now()})
		}
		//check id param
		id, err := strconv.Atoi(idstring)
		if err != nil {
			fmt.Println(id, err)
			return http.StatusNotAcceptable, err
		}
		//retrieve existing event for update
		event, err := c.repo.getEvent(id)
		if err != nil {
			fmt.Println(id, err)
			return http.StatusInternalServerError, errors.New("We could not find an event with the id specified.")
		}
		return c.renderTemplate(w, "new", event)
	} else if r.Method == "POST" {
		r.ParseForm()
		event := event{}
		event.Description = r.FormValue("description")
		event.Font = r.FormValue("font")
		idstring := r.FormValue("id")

		event.Date, _ = stringToTime(r.FormValue("date"))
		if idstring == "0"{
			fmt.Println("About to insert:", event)
			count, err := c.repo.insertEvent(&event)
			if err != nil || count == 0{
				fmt.Println(err)
				return c.renderTemplate(w, "new", event)
			}
		} else {
			event.ID, _ = strconv.Atoi(idstring)
			count, err := c.repo.updateEvent(&event)
			fmt.Println("About to update:", event)
			if err != nil || count == 0{
				fmt.Println(count, err)
				return c.renderTemplate(w, "new", event)
			}
		}
		http.Redirect(w, r, "/", http.StatusFound)
		return http.StatusFound, nil
	} else {
		return http.StatusMethodNotAllowed, errors.New("Bad Method")
	}
}
func EditHand(c *context, w http.ResponseWriter, r *http.Request) (int, error) {
	return 0, nil
}
func NotFoundHand(c *context, w http.ResponseWriter, r *http.Request) (int, error) {
	return http.StatusNotFound, errors.New("Not found.")
}
