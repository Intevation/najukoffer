package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/fcgi"
	"net/http/httputil"
	"os"
	"reflect"
	"runtime"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/cors"
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
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.Encode(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(buf.Bytes())
}

func respondWithGeoJSON(w http.ResponseWriter, code int, payload []termin) {
	featureCollection := geojson.NewFeatureCollection()
	for _, t := range payload {
		feature := geojson.NewPointFeature([]float64{t.X, t.Y})
		e := reflect.ValueOf(&t).Elem()
		for i := 0; i < e.NumField(); i++ {
			key := e.Type().Field(i).Name
			value := e.Field(i).Interface()
			if key != string("X") && key != string("Y") {
				feature.SetProperty(strings.ToLower(key), value)
			}
		}
		featureCollection.AddFeature(feature)
	}
	s, _ := featureCollection.MarshalJSON()
	//fmt.Println(string(s))
	w.Header().Set("Content-Type", "application/geo+json")
	w.WriteHeader(code)
	w.Write(s)
}

func getTermine(db *sql.DB, period string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		termine, err := getTermineFromDB(db, period)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithGeoJSON(w, http.StatusOK, termine)
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
		fmt.Sprintf("%s:%s@/%s?parseTime=true&collation=utf8mb4_general_ci&charset=utf8", dbUser, dbPasswd, dbName)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	mux := http.NewServeMux()

	if appAddr != "" {
		// Run as a local web server
		//mux.HandleFunc("/termine", isAuthorized(getTermine(db)))
		mux.HandleFunc("/today", getTermine(db, "today"))
		mux.HandleFunc("/this_week", getTermine(db, "this_week"))
		mux.HandleFunc("/this_month", getTermine(db, "this_month"))
		mux.HandleFunc("/this_year", getTermine(db, "this_year"))
		log.Println("Listening on " + appAddr + "...")
		//err = http.ListenAndServe(appAddr, requestLogger(mux))
		// cors.Default() setup the middleware with default options being
		// all origins accepted with simple methods (GET, POST). See
		// documentation below for more options.
		handler := cors.Default().Handler(mux)
		err = http.ListenAndServe(appAddr, handler)
	} else {
		// Run as FCGI via standard I/O
		mux.HandleFunc("/fcgi-bin/time.fcgi/termine", getTermine(db, "today"))
		err = fcgi.Serve(nil, mux)
	}
	if err != nil {
		log.Fatal(err)
	}

}
