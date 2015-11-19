package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type repo struct {
	db  *sqlx.DB
	log *log.Logger
}

var schema = `CREATE TABLE event (    
    id integer unique not null primary key,
    date text unique,
    title text,
    description text,
    font text
);`

func initDB(db *sqlx.DB) {
	db.MustExec(schema)
	tx := db.MustBegin()
	tx.MustExec("INSERT into event (date, title, description, font) values ($1, $2, $3, $4)", timeToString(time.Now()), "Do something fun!", "description goes here", "Helvetica")
	tx.MustExec("INSERT into event (date, title, description, font) values ($1, $2, $3, $4)", timeToString(time.Now().Add(time.Hour*24)), "Do something Else!", "discription goes here", "Helvetica")
	tx.Commit()
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
	r.log.Println("select * from event where date = ", date)
	row := r.db.QueryRowx("SELECT * FROM event where date = $1", date)
	err = e.mapRow(row)
	return
}
func (r repo) getEventsAfter(t time.Time) ([]event, error) {
	d := timeToString(t)
	rows, _ := r.db.Queryx("SELECT * FROM event where date >= $1", d)
	defer rows.Close()
	return mapEvents(rows)
}

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
	err := r.Scan(&e.ID, &temp,&e.Title, &e.Description, &e.Font)
	if err != nil {
		return err
	}
	e.Date, err = stringToTime(temp)
	return err //reutrn the error nil or not
}

func mapEvents(r *sqlx.Rows) ([]event, error) {
	events := make([]event, 0)
	for r.Next() {
		var temp string = ""
		e := event{}
		err := r.Scan(&e.ID, &temp,&e.Title, &e.Description, &e.Font)
		if err != nil {
			return nil, err
		}
		e.Date, err = stringToTime(temp)
		events = append(events, e)
	}
		/*doesn't work
		e := event{}
		err := e.mapRow(r)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
		*/					
	return events, nil
}

////////////////////
//      HELPERS   //
////////////////////

func timeToString(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
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
func getUID() int {
	t := time.Now()
	id, err := strconv.Atoi(fmt.Sprintf("%04d%02d%02d%02d%02d%02d", t.Year(), int(t.Month()), t.Day(), t.Hour()))
	if err != nil {
		return -1
	}
	return id
}

/*

 */
