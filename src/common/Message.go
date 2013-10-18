package common

import (
    "io/ioutil"
    "log"
    "net/http"
    "time"
)

type Message struct {
    Query string
    Engine string
    Url string
    User_Agent string
    Sleep_Time int
    //Time int64
}


func (m Message) GetHttpContent() (content string) {
    start_time := time.Now()

    client := &http.Client{}

    address := m.Url
    log.Println("address = " + address)

    req, err := http.NewRequest("GET", address, nil)
    if err != nil {
        log.Fatalln(err)
    }

    req.Header.Set("User-Agent", m.User_Agent)

    resp, err := client.Do(req)
    if err != nil {
        log.Fatalln(err)
        return "client do error"
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalln(err)
        return "read all error"
    }

    //log.Println(string(body))
    log.Println(time.Since(start_time))

    return (string(body))
}
