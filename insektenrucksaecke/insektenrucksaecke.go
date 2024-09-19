package insektenrucksaecke

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

var url = "https://cloud.naju.de/index.php/s/yfK9eoWP5m8ZYbE/download?path=%2FNAJU_Insektenrucksack_Ausleihstationen&files=NAJU_Ausleihstationen_Insektenrucksack.xlsx&downloadStartSecret=anltjezowne"

type insektenrucksack struct {
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

func getData(url string) ([]insektenrucksack, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	xlFile, err := xlsx.OpenBinary(body)
	if err != nil {
		log.Fatalln(err)
	}
	sheet := xlFile.Sheets[0]
	derInsektenrucksack := insektenrucksack{}
	var insektenrucksaecke []insektenrucksack
	rows := sheet.MaxRow
	for i := 0; i < rows; i++ {
		r, err := sheet.Row(i)
		if err != nil {
			return nil, err
		}
		if r.GetCell(9).Value == "" {
			continue
		}
		r.ReadStruct(&derInsektenrucksack)
		insektenrucksaecke = append(insektenrucksaecke, derInsektenrucksack)
	}
	return insektenrucksaecke, err
}

// Get : Get insektenrucksaecke
func Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		insektenrucksaecke, err := getData(url)
		if err != nil {
			helper.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		featureCollection := geojson.NewFeatureCollection()
		for _, insektenrucksack := range insektenrucksaecke {

			feature := geojson.NewPointFeature([]float64{insektenrucksack.Lon, insektenrucksack.Lat})

			e := reflect.ValueOf(&insektenrucksack).Elem()

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
