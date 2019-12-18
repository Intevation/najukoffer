package termine

import (
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/BjoernSchilberg/najukoffer/helper"
	geojson "github.com/paulmach/go.geojson"
)

type termin struct {
	Plz          string  `json:"plz"`
	Ort          string  `json:"ort"`
	Thema        string  `json:"thema"`
	Beschreibung string  `json:"beschreibung"`
	Von          string  `json:"von"`
	Bis          string  `json:"bis"`
	Bundesland   string  `json:"bundesland"`
	Kontakt      string  `json:"kontakt"`
	Kontaktdaten string  `json:"kontaktdaten"`
	Typ          string  `json:"typ"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
}

func getTermineFromDB(db *sql.DB, period string) ([]termin, error) {
	queryString :=
		fmt.Sprintf("SELECT CONVERT(plz,char(5)) as plz,ort,thema,beschreibung,von,bis,bundesland,eingetragen_von as kontakt,eingetragen_von_kontakt as kontaktdaten,typ,x,y FROM %s WHERE TYP REGEXP 'NAJU'", period)
	rows, err := db.Query(
		queryString,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	termine := []termin{}

	for rows.Next() {
		var t termin
		if err := rows.Scan(&t.Plz, &t.Ort, &t.Thema, &t.Beschreibung, &t.Von, &t.Bis, &t.Bundesland, &t.Kontakt, &t.Kontaktdaten, &t.Typ, &t.X, &t.Y); err != nil {
			return nil, err
		}
		termine = append(termine, t)
	}

	return termine, nil
}

func getNext6MonthFromDB(db *sql.DB) ([]termin, error) {
	queryString := "SELECT CONVERT(plz,char(5)) as plz,ort,thema,beschreibung,DATE_FORMAT(von,'%d.%m.%Y %H:%i') as von,DATE_FORMAT(bis,'%d.%m.%Y %H:%i') as bis,bundesland,eingetragen_von as kontakt, eingetragen_von_kontakt as kontaktdaten,typ,x,y FROM dates_with_location WHERE date(von) between curdate() and DATE_ADD(curdate(), INTERVAL 6 MONTH) and TYP REGEXP 'NAJU';"

	rows, err := db.Query(
		queryString,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	termine := []termin{}

	for rows.Next() {
		var t termin
		if err := rows.Scan(&t.Plz, &t.Ort, &t.Thema, &t.Beschreibung, &t.Von, &t.Bis, &t.Bundesland, &t.Kontakt, &t.Kontaktdaten, &t.Typ, &t.X, &t.Y); err != nil {
			return nil, err
		}
		termine = append(termine, t)
	}

	return termine, nil
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

// Get : Get dates from given period
func Get(db *sql.DB, period string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		termine, err := getTermineFromDB(db, period)
		if err != nil {
			helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithGeoJSON(w, http.StatusOK, termine)
	}
}

// GetNext6Month : Get dates from the coming next 6 month
func GetNext6Month(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		termine, err := getNext6MonthFromDB(db)
		if err != nil {
			helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithGeoJSON(w, http.StatusOK, termine)
	}
}
