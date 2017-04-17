package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"time"
	"encoding/json"
	"encoding/hex"
	"html/template"
	"io/ioutil"
	"net/http"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"goji.io/pat"
	"goji.io"
	"fmt"
)

type Stat struct {
	Hash        string `json:"device_hash"`
	Name        string `json:"device_name"`
	Version     string `json:"device_version"`
	Country     string `json:"device_country"`
	Carrier     string `json:"device_carrier"`
	CarrierId   string `json:"device_carrier_id"`
}

func (s Stat) Validate() error {
	//Verify hash
	//Hash should be non-empty & 64
	if s.Hash == "" {
		return errors.New("Invalid device hash")
	}
	if _, err := hex.DecodeString(s.Hash); err != nil {
		return errors.New("Device hash is invalid")
	}
	//Everything else should just be non-empty.
	if s.Name == "" {
		return errors.New("Invalid device model name")
	}
	if s.Version == "" {
		return errors.New("Invalid device version")
	}
	if s.Country == "" {
		return errors.New("Invalid deice country")
	}
	if s.Carrier == "" {
		return errors.New("Invalid device carrier")
	}
	if s.CarrierId == "" {
		return errors.New("Invalid device carrier ID")
	}
	return nil
}

func addStat(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, req * http.Request) {
		session := s.Copy()
		defer session.Close()

		statistic := session.DB("stats").C("statistic")
		aggregate := session.DB("stats").C("aggregate")


		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			//Failed reading body, this shouldn't happen ever.
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var s Stat
		err = json.Unmarshal(body, &s)

		if err != nil {
			//Internal unmarshal error. Shouldn't hit, even with malformed JSON
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = s.Validate()
		if err != nil {
			//Bad request (json malformed, bad data, etc)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = aggregate.Update(bson.M{"hash": s.Hash}, &s)
		if err != nil {
			//Failed to update aggregate. Falls into one of two cases:
			//1. Mongo problem. Bad data, bad data types, etc.
			//2. Doesn't exist. We need to create it.
			switch err {
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			case mgo.ErrNotFound:
				err = aggregate.Insert(&s)
				if err != nil {
					//We tried to create it, and failed. Falls back to #1 above.
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
		err = statistic.Insert(&s)
		if err != nil {
			///Mongo problem. Return error.
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}
}

type Thing struct {
	Name string
	Count int
}

type Context struct {
	LeftType string
	Left []Thing

	RightType string
	Right []Thing

	Count []int // [30 60 90]
}

func render_templates() error {

	var left []Thing
	var count []int
	left = append(left, Thing{"mako", 1})
	left = append(left, Thing{"shamu", 4})
	count = append(count, 3)
	count = append(count, 4)
	count = append(count, 5)
	context := &Context{LeftType: "model", Left: left, RightType: "Country", Right: left, Count: count}

	t := template.Must(template.New("main.html").ParseFiles("templates/main.html"))

	f, err := os.OpenFile("rendered/index.html", os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	err = t.Execute(f, context)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	portPtr := flag.Int("port", 8080, "port to launch on (default 8080")
	mUrlPtr := flag.String("mongo_url", "localhost", "Mongo url string (default localhost)")

	render_templates()
	session, err := mgo.Dial(*mUrlPtr)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/api/v1/stats"), addStat(session))

	//Please see README for explanation.
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Print("render")
				err := render_templates()
				if err != nil {
					log.Print(err)
				}

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	log.Print("Starting HTTP")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), nil))
}
