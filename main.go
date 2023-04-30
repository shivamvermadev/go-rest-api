package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pelletier/go-toml"
)

func getHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)
	w.Write([]byte("shivamverma"))
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)
	w.Write([]byte("shivam"))
}

func getPool(absPath string) (*sql.DB, error) {
	configTree, err := toml.LoadFile(absPath + "/main.toml")
	if err != nil {
		panic(err)
	}

	dbUsername := configTree.Get(fmt.Sprint("DATABASE", ".", "USERNAME")).(string)
	dbPassword := configTree.Get(fmt.Sprint("DATABASE", ".", "PASSWORD")).(string)
	dbName := configTree.Get(fmt.Sprint("DATABASE", ".", "NAME")).(string)
	dbHost := configTree.Get(fmt.Sprint("DATABASE", ".", "HOST")).(string)
	dbPort := configTree.Get(fmt.Sprint("DATABASE", ".", "PORT")).(string)

	db, err := sql.Open("mysql", fmt.Sprint(dbUsername, ":", dbPassword, "@(", dbHost, ":", dbPort, ")/", dbName)) // fmt.Sprint() will resolve to --> "root:test@(localhost:3306)/test"
	// db, err := sql.Open("mysql", "root:test@(localhost:3306)/test")

	if err != nil {
		panic(err)
	}

	if pe := db.Ping(); pe != nil {
		return nil, pe
	}

	db.SetConnMaxLifetime(time.Minute * 2)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}

func getConnection(ctx context.Context, dbPool *sql.DB) (*sql.Conn, error) {
	count := 0
	conn, conErr := dbPool.Conn(ctx)
	pingErr := conn.PingContext(ctx)

	for (conErr != nil || pingErr != nil) && count < 4 {
		//retry 4 times, actually 12 times (4 * 3, here by sql driver..check sql.go code)
		if conn != nil {
			conn.Close()
		}
		count = count + 1
		conn, conErr = dbPool.Conn(ctx)
		pingErr = conn.PingContext(ctx)
	}

	return conn, conErr
}

func main() {
	var dbPool *sql.DB
	var err error

	absPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	dbPool, err = getPool(absPath)

	if err != nil {
		panic(err)
	}

	conn, err := getConnection(context.Background(), dbPool)
	if err != nil {
		panic(err)
	}

	rows, _ := conn.QueryContext(context.Background(), "Select * from users")
	var name string
	var age int
	for rows.Next() {
		re := rows.Scan(&name, &age)
		if re != nil {
			continue
		}
		fmt.Println(name, age)
	}
	defer rows.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/", getHandler)

	mux.Handle("/test", http.HandlerFunc(testHandler))

	http.ListenAndServe(":8080", mux)
}
