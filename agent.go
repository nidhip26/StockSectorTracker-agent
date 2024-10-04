package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
    loggly "github.com/jamespearly/loggly"
)

type myStruct struct{
    Symbol             string    `json:"symbol"`
    Name               string    `json:"name"`
    Price              float64   `json:"price"`
    ChangesPercentage  float64   `json:"changesPercentage"`
    Change             float64   `json:"change"`
    DayLow             float64   `json:"dayLow"`
    DayHigh            float64   `json:"dayHigh"`
    YearHigh           float64   `json:"yearHigh"`
    YearLow            float64   `json:"yearLow"`
    MarketCap          int64     `json:"marketCap"`
    PriceAvg50         float64   `json:"priceAvg50"`
    PriceAvg200        float64   `json:"priceAvg200"`
    Exchange           string    `json:"exchange"`
    Volume             int64     `json:"volume"`
    AvgVolume          int64     `json:"avgVolume"`
    Open               float64   `json:"open"`
    PreviousClose      float64   `json:"previousClose"`
    EPS                float64   `json:"eps"`
    PE                 float64   `json:"pe"`

}

func fetchStocks(client *loggly.ClientType){
    c := http.Client{Timeout: time.Duration(1) * time.Second}                                     
    req, err := http.NewRequest("GET", "https://financialmodelingprep.com/api/v3/quote/AAPL,UNH,JPM,AMZN,PG,XOM,NEE,DOW,TMUS,AMT?apikey=bk5WxsGZsq2LrzdFBqQHu82YQOUv5Hkd", nil)
    if err != nil {
        fmt.Printf("error %s", err)
        message := fmt.Sprintf("Error creating request: %s", err)
        client.Send("error", message)
        return
    }
    req.Header.Add("Accept", `application/json`)
    resp, err := c.Do(req)
    if err != nil {
        fmt.Printf("error %s", err)
        message := fmt.Sprintf("Error during request: %s", err)
        client.Send("error", message)
        return
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        message := fmt.Sprintf("Error reading response body: %s", err)
        client.Send("error", message)
        return
    }
    //fmt.Printf("Body : %s", body)
    

    var myJSON []myStruct
    if err:= json.Unmarshal(body, &myJSON); err!=nil{
        message := fmt.Sprintf("Error unmarshaling JSON: %s", err)
        client.Send("error", message)
        return
    }
    message := fmt.Sprintf("Fetched %d records successfully.", len(myJSON))
    client.Send("info", message)
    fmt.Printf("Go Structure:\t%+v\n", myJSON)

}


func main() {

    var tag string
    tag = "agent"

    // Instantiate the client
    client := loggly.New(tag)

    interval := flag.Int("interval", 1, "Ticker interval in minutes")
    flag.Parse()    

    ticker := time.NewTicker(time.Duration(*interval) * time.Minute)
    defer ticker.Stop()

    fetchStocks(client)
    
    done := make(chan bool)
    go func() {
        for {
            select {
            case <-done:
                return
            case t := <-ticker.C:
                fmt.Println("Tick at", t)
                fetchStocks(client)
            }
      }
    }()

    select {}    
    

}
