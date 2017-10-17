package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"github.com/gorilla/mux"
	"strconv"
	"net/url"
	"errors"
	"encoding/json"
	"github.com/gorilla/handlers"
	"time"
)

type VariantResponse struct {
	Success bool `json:"success"`
	Variants []Variant `json:"variants"`
	Total []int `json:"total"`
	Error string `json:"error"`
}

type Variant struct {
	VariantId  string `json:"id"`
	Chromosome string `json:"chromosome"`
	Start int64 `json:"start"`
	Reference string `json:"reference"`
	Alternate string `json:"alternate"`
	DbSNP string `json:"dbSNP"`
	AC int64 `json:"AC"`
	AF JsonNullFloat64 `json:"AF"`
	NCalled int64 `json:"nCalled"`
	NNotCalled int64 `json:"nNotCalled"`
	NHomRef int64 `json:"nHomRef"`
	NHet int64 `json:"nHet"`
	NHomVar int64 `json:"nHomVar"`
	AltType string `json:"altType"`
	Consequence string `json:"consequence"`
	GeneMapping string `json:"geneMapping"`
	EXACMAF JsonNullString `json:"exacMAF"`
}

type VariantQuery struct {
	Chrom string
	Start int
	End int
	Lim int
	Skip int
}

func main() {
	var dsn string
	var connectionCheck bool
	flag.StringVar(&dsn, "d", "root:root@tcp(127.0.0.1:3306)/variants", "mysql dsn: [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]")
	flag.BoolVar(&connectionCheck, "u", false, "Connect to database and exit")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	db, err := sql.Open("mysql", dsn)
	check(err)
	defer db.Close()
	err = retry(10, 3 * time.Second,
		func() (error) {
			log.Println("connecting to mysql server...")
			return db.Ping()
		})
	check(err)
	fmt.Println("Connected to mysql server")
	if connectionCheck {
		return
	}

	r := mux.NewRouter().StrictSlash(true)
	r.Methods("GET, OPTIONS")
	r.HandleFunc("/", Index(r))
	r.HandleFunc("/variants", VariantsIndex(db))

	headersOk := handlers.AllowedHeaders([]string{"Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "OPTIONS"})
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(r)))
}

func retry(attempts int, sleep time.Duration, action func() (error)) (error) {
	var err error
	for i := 0; i < attempts; i++ {
		err = action()
		if err == nil {
			return err
		}
		time.Sleep(sleep)
		log.Println("retrying after error: ", err)
	}
	return err
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
		response := VariantResponse{true, []Variant{}, []int{}, ""}
		vq, err := parse(r.URL.Query())
		if err != nil {
			errorResponse(&response, err.Error())
			json.NewEncoder(w).Encode(response)
			return
		}

		count, err := countVariants(db, vq)
		if err != nil {
			errorResponse(&response, err.Error())
			json.NewEncoder(w).Encode(response)
			return
		}
		vs, err := queryVariants(db, vq, count)
		if err != nil {
			errorResponse(&response, err.Error())
			json.NewEncoder(w).Encode(response)
			return
		}
		response.Total = []int{count}
		response.Variants = vs
		json.NewEncoder(w).Encode(response)
	}
	return http.HandlerFunc(fn)
}

func parse (params url.Values) (VariantQuery, error) {
	errs := make([]string, 0)
	trackErrors := func(err error) {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	c, err := paramValue(params, "chromosome")
	trackErrors(err)
	start, err := parseNatural(params, "positionStart")
	trackErrors(err)
	end, err := parseNatural(params, "positionEnd")
	trackErrors(err)
	lim, err := parseNatural(params, "limit")
	if err != nil {
		lim = 500
	}
	skip, err := parseNatural(params, "skip")
	if err != nil {
		skip = 0
	}

	if len(errs) > 0 {
		return VariantQuery{}, errors.New(strings.Join(errs, ", "))
	}
	return VariantQuery{c, start, end, lim, skip}, nil
}

func countVariants(db *sql.DB, vq VariantQuery) (int, error) {
	count := "SELECT COUNT(*) AS size FROM vs where chromosome=? and start<=? and start>=?"
	rows, err := db.Query(count, vq.Chrom, vq.End, vq.Start)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		if err := rows.Scan(&n); err != nil {
			return -1, err
		}
	}
	if err := rows.Err(); err != nil || n <= 0 {
		return -1, err
	}
	return n, nil
}

func queryVariants(db *sql.DB, vq VariantQuery, count int) ([]Variant, error) {
	query := "SELECT * FROM vs where chromosome=? and start<=? and start>=? LIMIT ?, ?"
	variants, err := db.Query(query, vq.Chrom, vq.End, vq.Start, vq.Skip, vq.Lim)
	if err != nil {
		return nil, err
	}
	var size = count - vq.Skip
	if size < 0 {
		size = 0
	}
	if size > vq.Lim {
		size = vq.Lim
	}
	vs := make([]Variant, size)
	i := 0
	for variants.Next() {
		var v Variant
		//var ignored []byte
		if err := variants.Scan(&v.VariantId, &v.Chromosome, &v.Start, &v.Reference, &v.Alternate, &v.DbSNP, &v.AC, &v.AF, &v.NCalled, &v.NNotCalled, &v.NHomRef, &v.NHet, &v.NHomVar, &v.AltType); err != nil {
			return nil, err
		}
		vs[i] = v
		i++
	}
	if err := variants.Err(); err != nil {
		return nil, err
	}
	return vs, nil
}

func errorResponse(response *VariantResponse, err string) {
	response.Success = false
	response.Error = fmt.Sprintf("Errors: %s", err)
}

func paramValue(params url.Values, s string) (string, error) {
	if len(params) <= 0 || len(params[s]) <= 0 {
		return "", errors.New(fmt.Sprintf("Could not parse %s", s))
	}
	return params[s][0], nil
}

func parseNatural(params url.Values, s string) (int, error) {
	v, err := paramValue(params, s)
	if err != nil {
		return -1, err
	}
	vi, err := strconv.Atoi(v)
	if err != nil {
		return -1, err
	}

	if vi < 0 {
		return -1, errors.New(fmt.Sprintf("%s is less than 0", s))
	}
	return vi, nil
}

func check(err error) {
	if err == nil {
		return
	}
	panic(err)
}

type JsonNullFloat64 struct {
	sql.NullFloat64
}

type JsonNullString struct {
	sql.NullString
}

func (v JsonNullFloat64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Float64)
	} else {
		return json.Marshal(nil)
	}
}

func (v JsonNullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal(nil)
	}
}