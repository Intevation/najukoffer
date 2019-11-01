package landesverbaende

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/BjoernSchilberg/najukoffer/helper"
	geojson "github.com/paulmach/go.geojson"
	"github.com/tealeg/xlsx"
)

var url = "https://cloud.naju.de/index.php/s/HeAYwXkpz8skNnj/download?path=%2FNAJU_Gruppen_aktualisieren&files=Kindergruppen_Daten%20Website_05-2017.xlsx&downloadStartSecret=ldpuu0flwmj"

type landesverband struct {
	Landesverband string  `xlsx:"0"`
	Straße        string  `xlsx:"1"`
	Postleitzahl  string  `xlsx:"2"`
	Ort           string  `xlsx:"3"`
	Bundesland    string  `xlsx:"4"`
	Telefon       string  `xlsx:"5"`
	Fax           string  `xlsx:"6"`
	Mail          string  `xlsx:"7"`
	URL           string  `xlsx:"8"`
	Lat           float64 `xlsx:"9"`
	Lon           float64 `xlsx:"10"`
}

func getData(url string) ([]landesverband, error) {
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

	xlFile, error := xlsx.OpenBinary(body)
	if error != nil {
		log.Fatalln(error)
	}
	sheet := xlFile.Sheets[2]
	fmt.Println(sheet.Cols)
	derLandesverband := landesverband{}
	var landesverbaende []landesverband
	for i, row := range sheet.Rows {
		if i != 0 {
			if row != nil {
				row.ReadStruct(&derLandesverband)
				landesverbaende = append(landesverbaende, derLandesverband)
				//fmt.Printf("%+v\n", g)
			}
		}
	}
	//fmt.Printf("%v", target)

	return landesverbaende, error
}

// Get : Get kindergruppen
func Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		landesverbaende, err := getData(url)
		if err != nil {
			helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		featureCollection := geojson.NewFeatureCollection()
		for _, landesverband := range landesverbaende {

			feature := geojson.NewPointFeature([]float64{landesverband.Lon, landesverband.Lat})

			e := reflect.ValueOf(&landesverband).Elem()

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
