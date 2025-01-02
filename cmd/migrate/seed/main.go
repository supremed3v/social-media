package main

import (
	"github.com/supremed3v/social-media/internal/db"
	"github.com/supremed3v/social-media/internal/store"
)

func main() {
	conn, err := db.New("postgres://admin:adminpassword@localhost/social?sslmode=disable", 3, 3, "15m")

	if err != nil {
		return
	}
	defer conn.Close()
	store := store.NewStorage(conn)
	db.Seed(store)
}
