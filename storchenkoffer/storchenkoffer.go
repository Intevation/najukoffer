package storchenkoffer

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

var url = "https://cloud.naju.de/index.php/s/HeAYwXkpz8skNnj/download?path=%2FNAJU_Storchkoffer_Ausleihstation&files=Storchenkoffer_Ausleihstationen.xlsx&downloadStartSecret=67dcesra3mm"

type storchenkoffer struct {
	Name         string  `xlsx:"0"`
	Straße       string  `xlsx:"1"`
	Postleitzahl string  `xlsx:"2"`
	Ort          string  `xlsx:"3"`
	Bundesland   string  `xlsx:"4"`
	Telefon      string  `xlsx:"5"`
	URL          string  `xlsx:"6"`
	Lat          float64 `xlsx:"7"`
	Lon          float64 `xlsx:"8"`
}

func getData(url string) ([]storchenkoffer, error) {
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
	derStrochenkoffer := storchenkoffer{}
	var dieStorchenkoffer []storchenkoffer

	for i := 0; i < 1048576; i++ {
		r, err := sheet.Row(i)
		if err != nil {
			return nil, err
		}
		if r.GetCell(8).Value == "" {
			break
		}
		r.ReadStruct(&derStrochenkoffer)
		dieStorchenkoffer = append(dieStorchenkoffer, derStrochenkoffer)
	}
	//fmt.Printf("%v", target)

	return dieStorchenkoffer, err
}

// Get : Get all storchenkoffer
func Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dieStorchenkoffer, err := getData(url)
		if err != nil {
			helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		featureCollection := geojson.NewFeatureCollection()
		for _, derStorchenkoffer := range dieStorchenkoffer {

			feature := geojson.NewPointFeature([]float64{derStorchenkoffer.Lon, derStorchenkoffer.Lat})

			e := reflect.ValueOf(&derStorchenkoffer).Elem()

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
