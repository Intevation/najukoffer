package konsultationskitas

import (
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/BjoernSchilberg/najukoffer/helper"
	geojson "github.com/paulmach/go.geojson"
	"github.com/tealeg/xlsx/v3"
)

// var url = "https://cloud.naju.de/..." // TODO
// var url = "http://localhost:9080/konsultationskitas.xlsx" // TODO
var url = "http://mapserver.nabu.de:9080/konsultationskitas.xlsx" // TODO: this is expecting `python3 -m http.server 9080` to be serving test files

type konsultationskita struct {
	Name               string  `xlsx:"1"`
	Traeger            string  `xlsx:"2"`
	Strasse            string  `xlsx:"3"`
	Postleitzahl       string  `xlsx:"4"`
	Ort                string  `xlsx:"5"`
	Bundesland         string  `xlsx:"6"`
	Telefon            string  `xlsx:"7"`
	Mail               string  `xlsx:"8"`
	Lat                float64 `xlsx:"9"`
	Lon                float64 `xlsx:"10"`
	URL                string  `xlsx:"11"`
	Themenschwerpunkte string  `xlsx:"12"`
}

func getData(url string) ([]konsultationskita, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//log.Println(string(body))

	xlFile, err := xlsx.OpenBinary(body)
	if err != nil {
		log.Fatalln(err)
	}
	sheet := xlFile.Sheets[0]
	dieKonsultationskita := konsultationskita{}
	var konsultationskitas []konsultationskita
	rows := sheet.MaxRow
	for i := 0; i < rows; i++ {
		r, err := sheet.Row(i)
		if err != nil {
			return nil, err
		}
		if r.GetCell(9).Value == "" {
			continue
		}
		r.ReadStruct(&dieKonsultationskita)
		konsultationskitas = append(konsultationskitas, dieKonsultationskita)
	}
	return konsultationskitas, err
}

// Get : Get konsultationskitas
func Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		konsultationskitas, err := getData(url)
		if err != nil {
			helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		featureCollection := geojson.NewFeatureCollection()
		for _, konsultationskita := range konsultationskitas {

			feature := geojson.NewPointFeature([]float64{konsultationskita.Lon, konsultationskita.Lat})

			e := reflect.ValueOf(&konsultationskita).Elem()

			for i := 0; i < e.NumField(); i++ {
				key := e.Type().Field(i).Name
				value := e.Field(i).Interface()
				if key != string("Lat") && key != string("Lon") {
					feature.SetProperty(strings.ToLower(key), value)
				}
			}
			featureCollection.AddFeature(feature)
		}
		s, _ := featureCollection.MarshalJSON()
		//fmt.Println(string(s))
		w.Header().Set("Content-Type", "application/geo+json")
		w.WriteHeader(http.StatusOK)
		w.Write(s)

	}

}
