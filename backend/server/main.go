package main

import (
	"log"
	"database/sql"
)


type RegisterRequest{
    name string
	email string
	password string
}

func main() {
	log.Println("Start Main")
}



func openDB(){
    dbConStr := "host=" + os.Getenv("DB_HOST") +
		" port=" + os.Getenv("DB_PORT") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASS") +
		" dbname=" + os.Getenv("DB_NAME") +
		" sslmode=disable"
}
