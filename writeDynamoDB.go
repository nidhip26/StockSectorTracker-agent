package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
    "os"
    "context"
    loggly "github.com/jamespearly/loggly"
    "github.com/google/uuid"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type myStruct struct {
    Symbol             string  `json:"symbol"`
    Name               string  `json:"name"`
    Price              float64 `json:"price"`
    ChangesPercentage  float64 `json:"changesPercentage"`
    Change             float64 `json:"change"`
    DayLow             float64 `json:"dayLow"`
    DayHigh            float64 `json:"dayHigh"`
    YearHigh           float64 `json:"yearHigh"`
    YearLow            float64 `json:"yearLow"`
    MarketCap          int64   `json:"marketCap"`
    PriceAvg50         float64 `json:"priceAvg50"`
    PriceAvg200        float64 `json:"priceAvg200"`
    Exchange           string  `json:"exchange"`
    Volume             int64   `json:"volume"`
    AvgVolume          int64   `json:"avgVolume"`
    Open               float64 `json:"open"`
    PreviousClose      float64 `json:"previousClose"`
    EPS                float64 `json:"eps"`
    PE                 float64 `json:"pe"`
}

func saveToDynamoDB(stock myStruct) error {

    uuidVal := uuid.New().String()

    cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
        o.Region = "us-east-1"
        return nil
    })
    if err != nil {
        return err
    }

    svc := dynamodb.NewFromConfig(cfg)
    _, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
        TableName: aws.String("npatel4_stocks"),
        Item: map[string]types.AttributeValue{
            "id":                 &types.AttributeValueMemberS{Value: uuidVal},
            "symbol":             &types.AttributeValueMemberS{Value: stock.Symbol},
            "name":               &types.AttributeValueMemberS{Value: stock.Name},
            "price":              &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.Price)},
            "changesPercentage":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.ChangesPercentage)},
            "change":             &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.Change)},
            "dayLow":             &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.DayLow)},
            "dayHigh":            &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.DayHigh)},
            "yearHigh":           &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.YearHigh)},
            "yearLow":            &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.YearLow)},
            "marketCap":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", stock.MarketCap)},
            "priceAvg50":         &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.PriceAvg50)},
            "priceAvg200":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.PriceAvg200)},
            "exchange":           &types.AttributeValueMemberS{Value: stock.Exchange},
            "volume":             &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", stock.Volume)},
            "avgVolume":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", stock.AvgVolume)},
            "open":               &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.Open)},
            "previousClose":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.PreviousClose)},
            "eps":                &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.EPS)},
            "pe":                 &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", stock.PE)},
        },
    })

    return err
}

func fetchStocks(client *loggly.ClientType, apiKey string){
    c := http.Client{Timeout: time.Duration(1) * time.Second}
   // req, err := http.NewRequest("GET", "https://financialmodelingprep.com/api/v3/quote/AAPL,UNH,JPM,AMZN,PG,XOM,NEE,DOW,TMUS,AMT?apikey=bk5WxsGZsq2LrzdFBqQHu82YQOUv5Hkd", nil)
    reqURL := fmt.Sprintf("https://financialmodelingprep.com/api/v3/quote/AAPL,UNH,JPM,AMZN,PG,XOM,NEE,DOW,TMUS,AMT?apikey=%s", apiKey)
    req, err := http.NewRequest("GET", reqURL, nil)
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
    //fmt.Printf("Go Structure:\t%+v\n", myJSON)
    for _, stock := range myJSON {
        fmt.Printf("%+v\n", stock)
        if err := saveToDynamoDB(stock); err != nil {
           message := fmt.Sprintf("Failed to save stock %s to DynamoDB: %s\n", stock.Symbol, err)
           client.Send("error", message)
        }
    }
}

func main() {
    var tag string
    tag = "agent"

    // Instantiate the client
    client := loggly.New(tag)

    interval := flag.Int("interval", 1, "Ticker interval in minutes")
    flag.Parse()

    apiKey := os.Getenv("API_KEY")
    if apiKey == "" {
        fmt.Println("API_KEY environment variable is not set.")
        return
    }

    ticker := time.NewTicker(time.Duration(*interval) * time.Minute)
    defer ticker.Stop()

    fetchStocks(client, apiKey)

    done := make(chan bool)
    go func() {
        for {
            select {
            case <-done:
                return
            case t := <-ticker.C:
                fmt.Println("Tick at", t)
                fetchStocks(client, apiKey)
            }
        }
    }()

    select {}
}
