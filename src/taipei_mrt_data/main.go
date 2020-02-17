package main

import (
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "os"
    "encoding/json"
    "strings"
    iconv "github.com/djimenez/iconv-go"
)

// declare json format - map part
type Map struct {
    Lines []Line `json:"lines"`
}

type Line struct {
    Code string `json:"code"`
    Name string `json:"name"`
    Stations []Station `json:"stations"`
}

type Station struct {
    Code string `json:"code"`
    Name string `json:"name"`
}

// declare json format - train arrival data part
type Arrival struct {
    Train []Train `json:"resource"`
}

type Train struct {
    Station string `json:"Station"`
    Destination string `json:"Destination"`
    UpdateTime string `json:"UpdateTime"`
}


func main() {
    // mrt map json
    mrt_json_file := "taipei_mrt_station_json.json"
    mrt_json := ReadFile(mrt_json_file)

    var mrt_map Map
    json.Unmarshal([]byte(mrt_json), &mrt_map)

    // train arrival realtime data
    file_url := "http://tcgmetro.blob.core.windows.net/stationnames/stations.json"
    local_filename := "train_arrival.json"
    local_filename_utf8 := "train_arrival_utf8.json"

    if err := DownloadFile(local_filename, file_url); err != nil {
        panic(err)
    }

    Big5ToUTF8(local_filename, local_filename_utf8)

    train_data := ReadFile(local_filename_utf8)

    var arrival_data Arrival
    json.Unmarshal([]byte(train_data), &arrival_data)

    // loop line data
    var message string
    for i := 0; i < len(mrt_map.Lines); i++ {
        line := mrt_map.Lines[i];
        fmt.Println(line.Name);
        for j := 0; j < len(line.Stations); j++ {
            station := line.Stations[j];
            message = "- " + station.Name;
            // find arrival train for each station
            for k := 0; k < len(arrival_data.Train); k++ {
                train := arrival_data.Train[k];
                if station.Name == UnifyStationName(train.Station) {
                    message = message + " 往 " + UnifyStationName(train.Destination)
                }
            }
            // print result
            fmt.Println(message);
        }
    }

    /*
    for i := 0; i < len(arrival_data.Train); i++ {
        train := arrival_data.Train[i];
        fmt.Printf("%s 往 %s\n", UnifyStationName(train.Station), UnifyStationName(train.Destination));
    }
    */
}

func UnifyStationName(station_name string) string {
    new_station_name := strings.Replace(station_name, "站", "", -1)
    new_station_name = strings.Replace(new_station_name, "台北車", "台北車站", -1)
    return new_station_name
}

func Big5ToUTF8(in_filepath string, out_filepath string) {
    f, err := os.Open(in_filepath)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    reader, err := iconv.NewReader(f, "big5", "utf-8")
    if err != nil {
        panic(err)
    }

    fo, err := os.Create(out_filepath)
    if err != nil {
        panic(err)
    }
    defer fo.Close()

    io.Copy(fo, reader)
}

func ReadFile(filepath string) string {
    content, err := ioutil.ReadFile(filepath)
    if err != nil {
        panic(err)
    }
    text := string(content)
    return text
}

func DownloadFile(filepath string, url string) error {

    // get data
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // create file
    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    // write file
    _, err = io.Copy(out, resp.Body)
    return err
}
