package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mritunjaykumar/url-lookup-service/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type urlDataProvider struct {
	url      string
	database string
	session  *mgo.Session
}

func newURLDataProvider(args []string) *urlDataProvider {
	params := new(params)
	params.getParams(args[0])

	m := new(urlDataProvider)
	m.url = params.URL
	m.database = params.Mongodb
	log.Println("mongodb url:" + m.url)

	var err error
	m.session, err = mgo.Dial(m.url)

	if err != nil {
		log.Fatal(err)
	}

	return m
}

func (u *urlDataProvider) GetExistingMalwareUrls(w http.ResponseWriter, r *http.Request) {
	parameters := r.URL.Query()
	urls := parameters["url"]
	url := urls[0]
	db := u.session.DB(u.database)
	c := db.C("Malware")
	if c == nil {
		fmt.Println("collection is nil.")
	}

	var rec models.MalwareDataModel
	err := c.Find(bson.M{"url": url}).One(&rec)

	outputModel := new(models.MalwareOutputModel)
	outputModel.IsMalware = true
	if err != nil {
		if err.Error() == "not found" {
			outputModel.IsMalware = false
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&outputModel); err != nil {
		panic(err)
	}
}
