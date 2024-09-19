package projektpartner

import (
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/Intevation/najukoffer/helper"
	geojson "github.com/paulmach/go.geojson"
	"github.com/tealeg/xlsx/v3"
)

var url = "https://cloud.naju.de/index.php/s/yfK9eoWP5m8ZYbE/download?path=%2FKKN-Karte&files=projektpartner.xlsx&downloadStartSecret=bikemnr6r4j"

type projektpartner struct {
	Projektpartner     string  `xlsx:"1"`
	Strasse            string  `xlsx:"2"`
	Postleitzahl       string  `xlsx:"3"`
	Ort                string  `xlsx:"4"`
	Bundesland         string  `xlsx:"5"`
	Telefon            string  `xlsx:"6"`
	Mail               string  `xlsx:"7"`
	Lat                float64 `xlsx:"8"`
	Lon                float64 `xlsx:"9"`
	URL                string  `xlsx:"10"`
	Themenschwerpunkte string  `xlsx:"11"`
}

func getData(url string) ([]projektpartner, error) {
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
	derProjektpartner := projektpartner{}
	var projektpartners []projektpartner
	rows := sheet.MaxRow
	for i := 0; i < rows; i++ {
		r, err := sheet.Row(i)
		if err != nil {
			return nil, err
		}
		if r.GetCell(8).Value == "" {
			continue
		}
		r.ReadStruct(&derProjektpartner)
		projektpartners = append(projektpartners, derProjektpartner)
	}
	return projektpartners, err
}

// Get : Get kindergruppen
func Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projektpartners, err := getData(url)
		if err != nil {
			helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		featureCollection := geojson.NewFeatureCollection()
		for _, projektpartner := range projektpartners {

			feature := geojson.NewPointFeature([]float64{projektpartner.Lon, projektpartner.Lat})

			e := reflect.ValueOf(&projektpartner).Elem()

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
