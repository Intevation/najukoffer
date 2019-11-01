package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/fcgi"
	"net/http/httputil"
	"os"
	"runtime"

	"github.com/BjoernSchilberg/terminkoffer/kindergruppen"
	"github.com/BjoernSchilberg/terminkoffer/storchenkoffer"
	"github.com/BjoernSchilberg/terminkoffer/termine"
	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

var appAddr string
var dbHost string
var dbUser string
var dbName string
var dbPasswd string
var signingKey []byte

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	signingKey = []byte(os.Getenv("SIGNINGKEY"))
	appAddr = os.Getenv("APPADDR")
	dbHost = os.Getenv("DBHOST")
	dbName = os.Getenv("DBNAME")
	dbUser = os.Getenv("DBUSER")
	dbPasswd = os.Getenv("DBPASSWD")
}

func requestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Println(err)
		}
		//log.Println(string(requestDump))
		log.Println(fmt.Sprintf("%q", requestDump))
	})
}

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Token"] != nil {

			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return signingKey, nil
			})

			if err != nil {
				fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {

			fmt.Fprintf(w, "Not Authorized!")
		}
	})
}

func main() {
	var err error

	connectionString :=
		fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&collation=utf8mb4_general_ci&charset=utf8", dbUser, dbPasswd, dbHost, dbName)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	mux := http.NewServeMux()

	if appAddr != "" {
		// Run as a local web server
		//mux.HandleFunc("/termine", isAuthorized termine.GetTermine(db)))
		mux.HandleFunc("/today", termine.Get(db, "today"))
		mux.HandleFunc("/this_week", termine.Get(db, "this_week"))
		mux.HandleFunc("/this_month", termine.Get(db, "this_month"))
		mux.HandleFunc("/this_year", termine.Get(db, "this_year"))
		mux.HandleFunc("/next_6month", termine.GetNext6Month(db))
		mux.HandleFunc("/kindergruppen", kindergruppen.Get())
		mux.HandleFunc("/storchenkoffer", storchenkoffer.Get())
		log.Println("Listening on " + appAddr + "...")
		//err = http.ListenAndServe(appAddr, requestLogger(mux))
		// cors.Default() setup the middleware with default options being
		// all origins accepted with simple methods (GET, POST). See
		// documentation below for more options.
		handler := cors.Default().Handler(mux)
		err = http.ListenAndServe(appAddr, handler)
	} else {
		// Run as FCGI via standard I/O
		mux.HandleFunc("/fcgi-bin/terminkoffer/today", termine.Get(db, "today"))
		mux.HandleFunc("/fcgi-bin/terminkoffer/this_week", termine.Get(db, "this_week"))
		mux.HandleFunc("/fcgi-bin/terminkoffer/this_month", termine.Get(db, "this_month"))
		mux.HandleFunc("/fcgi-bin/terminkoffer/this_year", termine.Get(db, "this_year"))
		mux.HandleFunc("/fcgi-bin/terminkoffer/next_6month", termine.GetNext6Month(db))
		mux.HandleFunc("/fcgi-bin/kindergruppen", kindergruppen.Get())
		mux.HandleFunc("/fcgi-bin/storchenkoffer", storchenkoffer.Get())
		err = fcgi.Serve(nil, mux)
	}
	if err != nil {
		log.Fatal(err)
	}

}
