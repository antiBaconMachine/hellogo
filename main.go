package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
)

const CELSIUS_OFFSET = -273.15

type weatherProvider interface {
	temperature(city string) (float64, error)
}

type openWeatherMap struct{ApiKey string}
type weatherUnderground struct{ApiKey string}

func main() {
    http.HandleFunc("/hello/", hello)
    http.HandleFunc("/weather/", weather)
    http.ListenAndServe("localhost:8085", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
}

func weather(w http.ResponseWriter, r *http.Request) {
	ow := openWeatherMap{os.Getenv("OPENWEATHER_API_KEY")}
	wu := weatherUnderground{os.Getenv("WEATHER_UNDERGROUND_API_KEY")}
	city := strings.SplitN(r.URL.Path, "/", 3)[2]
   
    owData, err := ow.temperature(city)
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
    }

	wuData, err := wu.temperature(city)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(map[string]interface{}{
		"city": city,
		"ow": owData,
		"wu": wuData,
	})
}

func (w openWeatherMap) temperature(city string) (float64, error) {
    resp, err := http.Get(fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?APPID=%s&q=%s", w.ApiKey,  city))
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

    return d.Main.Kelvin + CELSIUS_OFFSET, nil
}

func (w weatherUnderground) temperature(city string) (float64, error) {
	resp, err := http.Get(fmt.Sprintf("http://api.wunderground.com/api/%s/conditions/q/%s.json", w.ApiKey, city))
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d struct {
		Observation struct {
			Celcius float64 `json:"temp_c"`
		} `json:"current_observation"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	return d.Observation.Celcius, nil
}

