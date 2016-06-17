package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
)

type weatherProvider interface {
	temperature(city string) (float64, error)
}

type openWeatherMap struct{}

func main() {
    http.HandleFunc("/hello/", hello)
    http.HandleFunc("/weather/", weather)
    http.ListenAndServe(":8085", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
}

func weather(w http.ResponseWriter, r *http.Request) {
	wp := openWeatherMap{}
	city := strings.SplitN(r.URL.Path, "/", 3)[2]
    
    data, err := wp.temperature(city)
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(map[string]interface{}{
		"city": city,
		"kelvin": data,
	})
}

func (w openWeatherMap) temperature(city string) (float64, error) {
    resp, err := http.Get(fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?APPID=%s&q=%s", os.Getenv("OPENWEATHER_API_KEY"),  city))
    if err != nil {
		return 0, err
    }

    defer resp.Body.Close()

    var d struct {
		Main struct {
			Kelvin float64 `json:"temp"`
		} `json:"main"`
	}

    if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
    }

    return d.Main.Kelvin, nil
}
   
