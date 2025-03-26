package praxisbeispiele

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

var url = "https://cloud.naju.de/public.php/dav/files/RPoqRCWJTkpXd3Q/KKN-Karte/Beispiele_guter_Praxis.xlsx"

type praxisbeispiel struct {
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
	Kokita             string  `xlsx:"12"`
	Themenschwerpunkte string  `xlsx:"13"`
}

func getData(url string) ([]praxisbeispiel, error) {
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
	dasPraxisbeispiel := praxisbeispiel{}
	var praxisbeispiele []praxisbeispiel
	rows := sheet.MaxRow
	for i := 0; i < rows; i++ {
		r, err := sheet.Row(i)
		if err != nil {
			return nil, err
		}
		if r.GetCell(9).Value == "" {
			continue
		}
		r.ReadStruct(&dasPraxisbeispiel)
		praxisbeispiele = append(praxisbeispiele, dasPraxisbeispiel)
	}
	return praxisbeispiele, err
}

// Get : Get praxisbeispiele
func Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		praxisbeispiele, err := getData(url)
		if err != nil {
			helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		featureCollection := geojson.NewFeatureCollection()
		for _, praxisbeispiel := range praxisbeispiele {

			feature := geojson.NewPointFeature([]float64{praxisbeispiel.Lon, praxisbeispiel.Lat})

			e := reflect.ValueOf(&praxisbeispiel).Elem()

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
