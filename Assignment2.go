package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "github.com/jmoiron/jsonq"
    "gopkg.in/mgo.v2/bson"
    "gopkg.in/mgo.v2"
    "github.com/julienschmidt/httprouter"
    "errors"
    "strings"
    
)

type Udata struct {
    Id bson.ObjectId `json:"id" bson:"_id"`
    Name string `json:"name" bson:"name"`
    Address string `json:"address" bson:"address"`
    City string `json:"city" bson:"city"`
    State string `json:"state" bson:"state"`
    Zip string `json:"zip" bson:"zip"`
    Coordinate struct {
        Lat float64 `json:"lat" bson:"lat"`
        Lng float64 `json:"lng" bson:"lng"`
    } `json:"coordinate" bson:"coordinate"`
}

//Create New Location 
func create(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var u Udata
    URL := "http://maps.google.com/maps/api/geocode/json?address="
    //transfer the data into the local object
    json.NewDecoder(req.Body).Decode(&u)

    //Randomly generated unique ID
    u.Id = bson.NewObjectId()

    URL = URL +u.Address+ " " + u.City + " " + u.State + " " + u.Zip+"&sensor=false"
    URL = strings.Replace(URL, " ", "+", -1)
    fmt.Println("URL "+ URL)

    //call to Google map API
    response, err := http.Get(URL)
    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }

    jq := jsonq.NewQuery(resp)
    status, err := jq.String("status")
    fmt.Println(status)
    if err != nil {
        return
    }
    if status != "OK" {
        err = errors.New(status)
        return
    }

    lat, err := jq.Float("results" ,"0","geometry", "location", "lat")
   if err != nil {
       fmt.Println(err)
        return
    }
    lng, err := jq.Float("results", "0","geometry", "location", "lng")
    if err != nil {
        fmt.Println(err)
        return
    }

    u.Coordinate.Lat = lat
    u.Coordinate.Lng = lng

    //store data into Mongo Lab
    newSession().DB("cmpe273").C("test").Insert(u)

    // interface to JSON struct
    reply, _ := json.Marshal(u)

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)

}

//Get a Location        
func get(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    uniqueid :=  p.ByName("uniqueid")

    if !bson.IsObjectIdHex(uniqueid) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(uniqueid)

    responseObj := Udata{}

    if err := newSession().DB("cmpe273").C("test").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }

    reply, _ := json.Marshal(responseObj)

    
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", reply)
}

//Update Location 
func update(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var u Udata
    uniqueid :=  p.ByName("uniqueid")

    URL := "http://maps.google.com/maps/api/geocode/json?address="

    //transfer the data into the local object
    json.NewDecoder(req.Body).Decode(&u)

    URL = URL +u.Address+ " " + u.City + " " + u.State + " " + u.Zip+"&sensor=false"
    URL = strings.Replace(URL, " ", "+", -1)
    fmt.Println("URL "+ URL)

    //Google map API
    response, err := http.Get(URL)
    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }

    jq := jsonq.NewQuery(resp)
    status, err := jq.String("status")
    fmt.Println(status)
    if err != nil {
        return
    }
    if status != "OK" {
        err = errors.New(status)
        return
    }

    lat, err := jq.Float("results" ,"0","geometry", "location", "lat")
    if err != nil {
        fmt.Println(err)
        return
    }
    lng, err := jq.Float("results", "0","geometry", "location", "lng")
    if err != nil {
        fmt.Println(err)
        return
    }

    u.Coordinate.Lat = lat
    u.Coordinate.Lng = lng

    dataid := bson.ObjectIdHex(uniqueid)
    var data = Udata{
        Address: u.Address,
        City: u.City,
        State: u.State,
        Zip: u.Zip,
    }
    //updatedata
    fmt.Println(data)
    //store data
    newSession().DB("cmpe273").C("test").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "address": u.Address,
        "city": u.City, "state": u.State,"zip": u.Zip, "coordinate.lat":u.Coordinate.Lat, "coordinate.lng":u.Coordinate.Lng}})

    responseObj := Udata{}

    //retrive the response data
    if err := newSession().DB("cmpe273").C("test").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }
    // interface into JSON struct
    reply, _ := json.Marshal(responseObj)

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)

}

//Delete a Location
func delete(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    uniqueid :=  p.ByName("uniqueid")

    if !bson.IsObjectIdHex(uniqueid) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(uniqueid)

    // delete user
    if err := newSession().DB("cmpe273").C("test").RemoveId(dataid); err != nil {
        rw.WriteHeader(404)
        return
    }

    rw.WriteHeader(200)
}

func newSession() *mgo.Session {
    //Connecting to Mongo Lab
    s, err := mgo.Dial("mongodb://dbuser:dbpswd@ds041934.mongolab.com:41934/cmpe273")

    // Check if mongo server running
    if err != nil {
        panic(err)
    }
    return s
}

func main() {
    mux := httprouter.New()
    mux.GET("/locations/:uniqueid", get)
    mux.POST("/locations", create)
    mux.PUT("/locations/:uniqueid", update)
    mux.DELETE("/locations/:uniqueid", delete)
        server := http.Server{
        Addr:        "0.0.0.0:5555",
        Handler: mux,
    }
    server.ListenAndServe()
}