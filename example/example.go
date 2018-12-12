package main

import (
	"log"
	"net/http"
	"database/sql"

	"github.com/apuigsech/rest-layer-sql"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"

	_ "github.com/mattn/go-sqlite3"
	//_ "github.com/gwenn/gosqlite"
)

const (
	DB_DRIVER		= "sqlite3"
	DB_SOURCE		= "file::memory:?cache=shared"

	DB_TABLE_UP		= "CREATE TABLE IF NOT EXISTS units (id VARCHAR(128) PRIMARY KEY,etag VARCHAR(128),updated TIMESTAMP,created TIMESTAMP,str VARCHAR(150),int INTEGER)"
)

var (
	unit = schema.Schema{
		Fields: schema.Fields{
			"id": schema.IDField,
			"created": schema.CreatedField,
			"updated": schema.UpdatedField,
			"str": {
				Sortable: true,
				Filterable: true,
				Required: true,
				Validator: &schema.String{
					MaxLen: 150,
				},
			},
			"int": {
				Sortable: true,
				Filterable: true,
				Required: true,
				Validator: &schema.Integer{},
			},
		},
	}
)

func main() {
	db, err := sql.Open(DB_DRIVER, DB_SOURCE)
	if err != nil {
		log.Fatalf("Invalid DB configuration: %s", err)
	}

	_, err = db.Exec(DB_TABLE_UP)
	if err != nil {
		log.Fatalf("Invalid DB configuration: %s", err)
	}	

	index := resource.NewIndex()

	index.Bind("units", unit, sqlStorage.NewHandler(db, "units"), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatalf("Invalid API configuration: %s", err)
	}

	http.Handle("/", api)

	log.Print("Serving API on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}