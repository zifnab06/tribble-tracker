package main

import (
	"errors"
	"log"
	"os"
	"time"
	"encoding/json"
	"encoding/hex"
	"html/template"
	"io/ioutil"
	"net/http"

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
	var err error
	//Verify hash
	//Hash should be non-empty & 64
	if s.Hash == "" {
		return errors.New("Invalid device hash")
	}
	if _, err = hex.DecodeString(s.Hash); err != nil {
		return errors.New("Device hash is invalid")
	}
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

func statsHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	//return req.Method
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var s Stat
	err = json.Unmarshal(body, &s)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = s.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	return
}

func mainHandler(w http.ResponseWriter, req *http.Request) {

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
	render_templates()
	http.HandleFunc("/api/v1/stats", statsHandler)
	http.HandleFunc("/", mainHandler)
	log.Print("foobar")
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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
