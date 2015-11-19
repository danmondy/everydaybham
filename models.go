package main

import (
	//"fmt"
	"time"
)

type event struct {
	ID          int
	Title       string
	Date        time.Time
	Description string
	Font        string
}

type user struct {
	ID   int
	Date time.Time
}

type session struct {
	Date   time.Time
	UserID int
}
