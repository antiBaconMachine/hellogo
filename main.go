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

type multiWeatherProvider []weatherProvider

type openWeatherMap struct{ApiKey string}
type weatherUnderground struct{ApiKey string}

var wp multiWeatherProvider

func main() {
	wp = multiWeatherProvider {
		openWeatherMap{os.Getenv("OPENWEATHER_API_KEY")},
		weatherUnderground{os.Getenv("WEATHER_UNDERGROUND_API_KEY")},
	}
    http.HandleFunc("/hello/", hello)
    http.HandleFunc("/weather/", weather)
    http.ListenAndServe(":8085", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello"))
}

func weather(w http.ResponseWriter, r *http.Request) {
	city := strings.SplitN(r.URL.Path, "/", 3)[2]
   
    average, err := wp.temperature(city)
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(map[string]interface{}{
		"city": city,
		"celcius": average,
	})
}

func (providers multiWeatherProvider) temperature(city string) (float64, error) {
	temps := make(chan float64, len(providers))
	errs := make(chan error, len(providers))

	for _, provider := range providers {
		go func(provider weatherProvider) {
			c, err := provider.temperature(city)
			if err != nil {
				errs <- err
				return 
			}
			temps <- c
		}(provider)
	}

	sum := 0.0
	
	for i := 0; i< len(providers); i++ {
		select {
		case temp := <-temps:
			sum += temp
		case err := <-errs:
			return 0, err
		}
	}

	return sum / float64(len(providers)), nil
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

