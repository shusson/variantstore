package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"github.com/gorilla/mux"
)

func main() {
	var dsn string
	flag.StringVar(&dsn, "d", "root:root@tcp(127.0.0.1:3306)/v", "mysql dsn: [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	db, err := sql.Open("mysql", dsn)
	check(err)
	defer db.Close()
	err = db.Ping()
	check(err)

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", Index(r))
	r.HandleFunc("/variants", VariantsIndex(db))
	r.HandleFunc("/variants/{variantId}", VariantsShow)

	log.Fatal(http.ListenAndServe(":8080", r))
}

func Index(router *mux.Router) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Simple variant store API")
		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			t, err := route.GetPathTemplate()
			if err != nil {
				return err
			}
			fmt.Fprintln(w, t)
			return nil
		})
	}
	return http.HandlerFunc(fn)
}

func VariantsIndex(db *sql.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "VARIANTS!")
	}
	return http.HandlerFunc(fn)
}

func VariantsShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	variantId := vars["variantId"]
	fmt.Fprintln(w, "variant:", variantId)
}

func loadData(file string, Db *sql.DB) {
	basename := filepath.Base(file)
	name := strings.TrimSuffix(basename, filepath.Ext(file))
	query := fmt.Sprintf("LOAD DATA INFILE '/var/lib/mysql-files/%s' INTO TABLE `%s`", basename, name)
	fmt.Printf("Loading %s\n", name)
	_, err := Db.Exec(query)
	check(err)
}

func check(err error) {
	if err == nil {
		return
	}
	panic(err)
}