package main

import (
	"log"
)

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal("veritabaninda hata", err)
	}

	defer store.Close() // Veritabanı bağlantısını kapatır.

	if err, err2 := store.Init(); err != nil {
		log.Fatal(err)
	} else if err2 != nil {
		log.Fatal(err2)
	}
	server := NewApiServer(":8080", store)
	server.Run()
}
