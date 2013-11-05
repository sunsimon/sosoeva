package common

import (
    "io/ioutil"
    "log"
    "net/http"
    "time"
    "os/exec"
    "strings"
    "bytes"
)

type Message struct {
    Query string
    Engine string
    Url string
    User_Agent string
    Sleep_Time int
    Is_Use_Phantomjs bool
}

func (m Message) GetContent()(content string) {
    if m.Is_Use_Phantomjs {
        return m.GetContentByPhantomjs()
    } else {
        return m.GetHttpContent()
    }
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

func (m Message) GetContentByPhantomjs()(content string) {
    cmd := exec.Command("./tools/phantomjs", "./tools/download.js",  m.Url, m.User_Agent)
    cmd.Stdin = strings.NewReader("some input")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        log.Println(err)
        return "000"
    }
    return out.String()
}
