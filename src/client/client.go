package main

import (
    "log"
    //"os"
    "time"
    "common"
    "ssn/ssn_utils"
    "ssn/ssn_config"
    "deliver"
    "strconv"
)

func main() {
    start_time := time.Now()

    read_query_file := "query.txt"
    lines, err := common.ReadLinesFromFile(read_query_file)
    ssn_utils.CheckError(err, "read query from file error")

    query_num := len(lines)
    log.Printf("query_num=%d", query_num)

    c := ssn_config.NewConfig("autoeva.conf")
    cerr := c.Read()
    log.Println(cerr)
    user_agent := c.Get("A", "USER_AGENT")

    soso_line := make(chan int)
    baidu_line := make(chan int)
    sogou_line := make(chan int)

    if c.Get("B", "soso") == "1" {
        m_soso := new (common.Message)
        m_soso.Query = "";
        m_soso.Engine = "soso"
        m_soso.Url = c.Get("A", "SOSO_URL")
        m_soso.User_Agent = user_agent
        soso_sleeptime, _ := strconv.Atoi(c.Get("B", "soso_sleeptime"))
        m_soso.Sleep_Time = soso_sleeptime 
        go deliver.Deliver(soso_line, m_soso, lines)
    }

    if c.Get("B", "baidu") == "1" {
        m_baidu := new (common.Message)
        m_baidu.Query = "";
        m_baidu.Engine = "baidu"
        m_baidu.Url = c.Get("A", "BAIDU_URL")
        m_baidu.User_Agent = user_agent
        baidu_sleeptime, _ := strconv.Atoi(c.Get("B", "baidu_sleeptime"))
        m_baidu.Sleep_Time = baidu_sleeptime
        go deliver.Deliver(baidu_line, m_baidu, lines)

    }

    if c.Get("B", "sogou") == "1" {
        m_sogou := new (common.Message)
        m_sogou.Query = "";
        m_sogou.Engine = "sogou"
        m_sogou.Url = c.Get("A", "SOGOU_URL")
        m_sogou.User_Agent = user_agent
        sogou_sleeptime, _ := strconv.Atoi(c.Get("B", "sogou_sleeptime"))
        m_sogou.Sleep_Time = sogou_sleeptime

        go deliver.Deliver(sogou_line, m_sogou, lines)
    }

    break_flag := false
    suc_count := 0
    for {
        select {
        case <-soso_line :
            log.Println("main process. soso done. Using time :", time.Since(start_time))
            suc_count ++ 
        case <-baidu_line :
            log.Println("main process. baidu done. Using time :", time.Since(start_time))
            suc_count ++ 
        case <-sogou_line :
            log.Println("main process. sogou done. Using time :", time.Since(start_time))
            suc_count ++
        case  <-time.After(time.Second * 10) :
            log.Println("main process. time out  break. Using time :", time.Since(start_time))
            break_flag = true
            break
        }
        if break_flag || suc_count == 3{
            break
        }
    }
    log.Println("All done... Use time :", time.Since(start_time))

}

