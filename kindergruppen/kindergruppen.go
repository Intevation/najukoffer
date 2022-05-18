package kindergruppen

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

var url = "https://cloud.naju.de/index.php/s/HeAYwXkpz8skNnj/download?path=%2FNAJU_Gruppen_aktualisieren&files=Kindergruppen_Daten%20Website_05-2017.xlsx&downloadStartSecret=ldpuu0flwmj"

type kindergruppe struct {
	Stadtverband    string  `xlsx:"0"`
	Gruppenname     string  `xlsx:"1"`
	Strasse         string  `xlsx:"2"`
	PLZ             string  `xlsx:"3"`
	ORT             string  `xlsx:"4"`
	AlterTN         string  `xlsx:"5"`
	Treffpunkt      string  `xlsx:"6"`
	Zeit            string  `xlsx:"7"`
	Webseite        string  `xlsx:"8"`
	Ansprechpartner string  `xlsx:"9"`
	EMail           string  `xlsx:"10"`
	Telefon         string  `xlsx:"11"`
	Lat             float64 `xlsx:"12"`
	Lon             float64 `xlsx:"13"`
}

func getData(url string) ([]kindergruppe, error) {
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
	sheet := xlFile.Sheets[2]
	dieKindergruppe := kindergruppe{}
	var kindergruppen []kindergruppe
	rows := sheet.MaxRow
	log.Println(rows)
	for i := 0; i < rows; i++ {
		r, err := sheet.Row(i)
		if err != nil {
			return nil, err
		}
		if r.GetCell(13).Value == "" {
			continue
		}
		r.ReadStruct(&dieKindergruppe)
		kindergruppen = append(kindergruppen, dieKindergruppe)
	}
	return kindergruppen, err
}

// Get : Get kindergruppen
func Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kindergruppen, err := getData(url)
		if err != nil {
			helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		featureCollection := geojson.NewFeatureCollection()
		for _, kindergruppe := range kindergruppen {

			feature := geojson.NewPointFeature([]float64{kindergruppe.Lon, kindergruppe.Lat})

			e := reflect.ValueOf(&kindergruppe).Elem()

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
