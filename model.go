package main

import "database/sql"

type termin struct {
	Plz          string  `json:"plz"`
	Ort          string  `json:"ort"`
	Thema        string  `json:"thema"`
	Beschreibung string  `json:"beschreibung"`
	Von          string  `json:"von"`
	Bis          string  `json:"bis"`
	Bundesland   string  `json:"bundesland"`
	Typ          string  `json:"typ"`
	X            float32 `json:"x"`
	Y            float32 `json:"y"`
}

func getTermineFromDB(db *sql.DB) ([]termin, error) {
	rows, err := db.Query(
		"SELECT CONVERT(plz,char(5)) as plz,ort,thema,beschreibung,von,bis,bundesland,typ,x,y FROM today",
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	termine := []termin{}

	for rows.Next() {
		var t termin
		if err := rows.Scan(&t.Plz, &t.Ort, &t.Thema, &t.Beschreibung, &t.Von, &t.Bis, &t.Bundesland, &t.Typ, &t.X, &t.Y); err != nil {
			return nil, err
		}
		termine = append(termine, t)
	}

	return termine, nil
}
