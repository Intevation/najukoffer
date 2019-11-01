package storchenkoffer

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/BjoernSchilberg/terminkoffer/helper"
	geojson "github.com/paulmach/go.geojson"
	"github.com/tealeg/xlsx"
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

	xlFile, error := xlsx.OpenBinary(body)
	if error != nil {
		log.Fatalln(error)
	}
	sheet := xlFile.Sheets[2]
	fmt.Println(sheet.Cols)
	derStrochenkoffer := storchenkoffer{}
	var dieStorchenkoffer []storchenkoffer
	for i, row := range sheet.Rows {
		if i != 0 {
			if row != nil {
				row.ReadStruct(&derStrochenkoffer)
				dieStorchenkoffer = append(dieStorchenkoffer, derStrochenkoffer)
				//fmt.Printf("%+v\n", g)
			}
		}
	}
	//fmt.Printf("%v", target)

	return dieStorchenkoffer, error
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
