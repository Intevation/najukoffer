package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/fcgi"
	"net/http/httputil"
	"os"
	"runtime"

	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var appAddr string
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

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func getTermine(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		termine, err := getTermineFromDB(db)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, termine)
	}
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
		fmt.Sprintf("%s:%s@/%s", dbUser, dbPasswd, dbName)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	mux := http.NewServeMux()

	if appAddr != "" {
		// Run as a local web server
		mux.HandleFunc("/termine", isAuthorized(getTermine(db)))
		log.Println("Listening on " + appAddr + "...")
		//	err = http.ListenAndServe(appAddr, mux)
		err = http.ListenAndServe(appAddr, requestLogger(mux))
	} else {
		// Run as FCGI via standard I/O
		mux.HandleFunc("/fcgi-bin/time.fcgi/termine", getTermine(db))
		err = fcgi.Serve(nil, mux)
	}
	if err != nil {
		log.Fatal(err)
	}

}
