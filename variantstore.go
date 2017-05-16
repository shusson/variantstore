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
	"strconv"
	"net/url"
	"errors"
	"encoding/json"
	"github.com/gorilla/handlers"
)

type VariantResponse struct {
	Success bool `json:"success"`
	Variants []Variant `json:"variants"`
	Total []int `json:"total"`
	Error string `json:"error"`
}

type Variant struct {
	Chromosome string `json:"chromosome"`
	Start int64 `json:"start"`
	Reference string `json:"reference"`
	Alternate string `json:"alternate"`
	DbSNP string `json:"dbSNP"`
	CallRate float64 `json:"callRate"`
	AC int64 `json:"AC"`
	AF float64 `json:"AF"`
	NCalled int64 `json:"nCalled"`
	NNotCalled int64 `json:"nNotCalled"`
	NHomRef int64 `json:"nHomRef"`
	NHet int64 `json:"nHet"`
}

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
	r.Methods("GET, OPTIONS")
	r.HandleFunc("/", Index(r))
	r.HandleFunc("/variants", VariantsIndex(db))
	r.HandleFunc("/variants/{variantId}", VariantsShow)

	headersOk := handlers.AllowedHeaders([]string{"Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "OPTIONS"})
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(r)))
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

func errorResponse(response *VariantResponse, err string) {
	response.Success = false
	response.Error = fmt.Sprintf("Errors: %s", err)
}

func VariantsIndex(db *sql.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		errs := make([]string, 0)
		trackErrors := func(err error) {
			if err != nil {
				errs = append(errs, err.Error())
			}
		}

		params := r.URL.Query()
		c, err := getParamValue(params, "chromosome")
		trackErrors(err)
		start, err := parseInt(params, "positionStart")
		trackErrors(err)
		end, err := parseInt(params, "positionEnd")
		trackErrors(err)
		lim, err := parseInt(params, "limit")
		if err != nil {
			lim = 500
		}
		skip, err := parseInt(params, "skip")
		if err != nil {
			skip = 0
		}

		response := VariantResponse{true, []Variant{}, []int{}, ""}

		if len(errs) > 0 {
			errorResponse(&response, strings.Join(errs, ", "))
			json.NewEncoder(w).Encode(response)
			return
		}

		count := "SELECT COUNT(*) AS size FROM vs where chromosome=? and start<=? and start>=?"
		rows, err := db.Query(count, c, end, start)
		if err != nil {
			errorResponse(&response, err.Error())
			json.NewEncoder(w).Encode(response)
			return
		}
		defer rows.Close()
		n := 0
		for rows.Next() {
			if err := rows.Scan(&n); err != nil {
				errorResponse(&response, err.Error())
				json.NewEncoder(w).Encode(response)
				return
			}
		}
		if err := rows.Err(); err != nil || n <= 0 {
			errorResponse(&response, err.Error())
			json.NewEncoder(w).Encode(response)
			return
		}

		query := "SELECT * FROM vs where chromosome=? and start<=? and start>=? LIMIT ?, ?"

		variants, err := db.Query(query, c, end, start, skip, lim)
		if err != nil {
			errorResponse(&response, err.Error())
			json.NewEncoder(w).Encode(response)
			return
		}

		vs := make([]Variant, n)
		i := 0
		for variants.Next() {
			var v Variant
			var ignored []byte
			if err := variants.Scan(&v.Chromosome, &v.Start, &v.Reference, &v.Alternate, &v.DbSNP, &v.CallRate, &v.AC, &v.AF, &v.NCalled, &v.NNotCalled, &v.NHomRef, &v.NHet, &ignored, &ignored, &ignored, &ignored, &ignored, &ignored, &ignored, &ignored, &ignored, &ignored); err != nil {
				errorResponse(&response, err.Error())
				json.NewEncoder(w).Encode(response)
				return
			}
			vs[i] = v
			i++
		}
		if err := variants.Err(); err != nil || n <= 0 {
			errorResponse(&response, err.Error())
			json.NewEncoder(w).Encode(response)
			return
		}
		response.Total = []int{n}
		response.Variants = vs
		json.NewEncoder(w).Encode(response)
	}
	return http.HandlerFunc(fn)
}

func VariantsShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	variantId := vars["variantId"]
	fmt.Fprintln(w, "variant:", variantId)
}

func getParamValue(params url.Values, s string) (string, error) {
	if len(params) <= 0 || len(params[s]) <= 0 {
		return "", errors.New(fmt.Sprintf("Could not parse %s", s))
	}
	return params[s][0], nil
}

func parseInt(params url.Values, s string) (int, error) {
	vs, err := getParamValue(params, s)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(vs)
}

func check(err error) {
	if err == nil {
		return
	}
	panic(err)
}