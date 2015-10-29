package main

import (
	"fmt"
	"time"
	"strings"
	"strconv"
	"errors"
	
	"github.com/jmoiron/sqlx"
)

type repo struct {
	db *sqlx.DB
}

/////////////////////
//      EVENTS     //
/////////////////////
func (r repo) getEvent(id int) (e event, err error) {
	e = event{}
	row := r.db.QueryRowx("SELECT * FROM event where id = $1", id)
	err = e.mapRow(row)
	return
}
func (r repo) getEventByDate(date string) (e event, err error) {
	e = event{}
	row := r.db.QueryRowx("SELECT * FROM event where date = $1", date)
	err = e.mapRow(row)
	return
}

//func (r repo) getEvents(time string)
func (r repo) insertEvent(e *event) (int64, error) {

	res, err := r.db.Exec("INSERT into event (date, description, font) VALUES ($1, $2, $3)", timeToString(e.Date), e.Description, e.Font)
	if err != nil {
		r, _ := res.RowsAffected()
		return r, err
	}
	if id, err := res.LastInsertId(); err != nil {
		e.ID = int(id)
		return res.LastInsertId()
	}
	return res.RowsAffected()
}
func (r repo) updateEvent(e *event) (int64, error) {
	result, err := r.db.Exec("UPDATE event set date=$1, description=$2, font=$3 where id=$4", timeToString(e.Date), e.Description, e.Font, e.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
func (e *event) mapRow(r *sqlx.Row) error {
	var temp string = ""
	err := r.Scan(&e.ID, &temp, &e.Description, &e.Font)
	if err != nil {
		return err
	}
	e.Date, err = stringToTime(temp)
	return err //reutrn the error nil or not
}

////////////////////
//      HELPERS   //
////////////////////

func timeToString(t time.Time) string {
	return fmt.Sprintf("%4d-%2d-%2d", t.Year(), t.Month(), t.Day())
	//time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	//return t.Format(time.RFC822)
}
func stringToTime(s string) (time.Time, error) {
	ymd := strings.Split(s, "-")
	y, err := strconv.Atoi(ymd[0])
	m, err1 := strconv.Atoi(ymd[1])
	d, err2 := strconv.Atoi(ymd[2])
	if err != nil || err1 != nil || err2 != nil {
		return time.Time{}, errors.New("Could not convert")
	}
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC), nil
	//return time.Parse(time.RFC822, s)
}

/*
func MapUsers(r *sql.Rows)([]User, error){
	var users []User
	for r.Next(){
		var u User
		err := r.Scan(&u.Id, &u.Email, &u.Hashword, &u.Rank, &u.Since)
		if err != nil{
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
*/
