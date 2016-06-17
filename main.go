package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
)

type weatherData struct {
    Name string `json:"name"`
    Main struct {
	Kelvin float64 `json:"temp"`
    } `json:"main"`
}

func main() {
    http.HandleFunc("/hello/", hello)
    http.HandleFunc("/weather/", weather)
    http.ListenAndServe(":8085", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
}

func weather(w http.ResponseWriter, r *http.Request) {
    city := strings.SplitN(r.URL.Path, "/", 3)[2]
    
    data, err := query(city)
    if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return 
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(data)
}

func query(city string) (weatherData, error) {
    resp, err := http.Get(fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?APPID=%s&q=%s", os.Getenv("OPENWEATHER_API_KEY"),  city))
    if err != nil {
	return weatherData{}, err
    }

    defer resp.Body.Close()

    var d weatherData

    if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
	return weatherData{}, err
    }

    return d, nil
}
   
