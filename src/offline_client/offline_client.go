package main

import proto "code.google.com/p/goprotobuf/proto"
import (
    "soso_proto"
    "database/sql"
    "log"
    "net"
    "flag"
    "os"
    "ssn/ssn_utils"
    "io"
    "time"
    _ "mysql"
)
//rpc use 
import "rpc_proto"
//for get query
import "strings"
import "strconv"

//Global Var
var server_ip *string
var parse_server_ip *string
var tid *string

const USER_AGENT = "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/32.0.1664.3"

func init(){
    //set flag
    server_ip = flag.String("server_ip", "10.11.195.131:9998", "download page  server ip")
    parse_server_ip = flag.String("parse_server_ip", "10.136.30.20:30026", "parse the html page  server ip")
    tid = flag.String("tid", "0", "task_id")
    flag.PrintDefaults()
    flag.Parse()

}

func main(){
    if *tid == "0" {
        log.Fatalf("must input the tid")
    }
    //get engine
    sql_string := "SELECT tid, engine FROM tb_dcg_task where tid=" + *tid
    result := GetResult(sql_string)
    log.Println(result)
    if len(result) != 1{
        log.Fatalf("no tid in tb_dcg_task")
    }
    
    engine_str, err := result[0]["engine"]
    if err != true {
        log.Fatalf("read engine in tb_dcg_task err")
    }

    engine_arr := strings.SplitAfter(engine_str, "\t")
    sql_string = "SELECT eid, name, url from tb_engine "
    sql_string += "where eid in (" + strings.Join(engine_arr, ",") + ")"
    engine_cfg := GetResult(sql_string)
    log.Println(engine_cfg)

    sql_string = "SELECT qid, query from tb_dcg_query "
    sql_string += "where tid=" + *tid
    sql_string += " limit 10"
    result = GetResult(sql_string)
    log.Println(result)

    //here linar download
    for _, item := range result {
        for _, engine_item := range engine_cfg {
            engine_name := engine_item["name"]
            crawl_url := strings.Replace(engine_item["url"], "__QUERY__", ssn_utils.UrlEncode(item["query"]), 1)
            crawl_query := item["query"]

            html_string := GetContent(crawl_query, engine_name, crawl_url)
            log.Println("GetContent:\n", html_string)
            parse_res := ParseHtml(engine_name, html_string)

            for i_rank, s_url := range parse_res {
                log.Println("rank=", i_rank)
                log.Println("url=", s_url)


                sql_string = "insert into tb_dcg_data_list set "
                sql_string += " task_id=" + *tid + ", query='" + crawl_query +"', "
                sql_string += " url='" + s_url + "', rank=" + strconv.Itoa(i_rank) + ", "
                sql_string += " engine=" + engine_item["eid"]
                ExecSQL(sql_string)
            }
        }
    }
}

func GetResult(sql_string string) (res map[int]map[string]string) {
    log.Println("sql:", sql_string)

    // Open database connection
    db, err := sql.Open("mysql", "jianwensun:sun@tcp(10.11.195.131:3306)/mark_oss")
    if err != nil {
        panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
    }
    defer db.Close()

    // Execute the query
    //rows, err := db.Query("SELECT tid, engine FROM tb_dcg_task")
    rows, err := db.Query(sql_string)
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }

    // Get column names
    columns, err := rows.Columns()
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }

    // Make a slice for the values
    values := make([]sql.RawBytes, len(columns))

    // rows.Scan wants '[]interface{}' as an argument, so we must copy the
    // references into such a slice
    // See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
    scanArgs := make([]interface{}, len(values))
    for i := range values {
        scanArgs[i] = &values[i]
    }

    result := make(map[int]map[string]string)
    count := 0
    // Fetch rows
    for rows.Next() {
        // get RawBytes from data
        err = rows.Scan(scanArgs...)
        if err != nil {
            log.Println("in rows.Scan")
            panic(err.Error()) // proper error handling instead of panic in your app
        }

        temp_store := make(map[string]string, len(columns))
        // Now do something with the data.
        // Here we just print each column as a string.
        var value string
        for i, col := range values {
            // Here we can check if the value is nil (NULL value)
            if col == nil {
                value = "NULL"
            } else {
                value = string(col)
            }
            log.Println(columns[i], ": ", value)
            temp_store[columns[i]] = value
        }
        result[count] = temp_store
        count ++
        log.Println("-----------------------------------")
    }
    return result 
}

func ExecSQL(sql_string string) {
    log.Println("Execsql:", sql_string)

    // Open database connection
    db, err := sql.Open("mysql", "jianwensun:sun@tcp(10.11.195.131:3306)/mark_oss")
    if err != nil {
        panic(err.Error())  
    }
    defer db.Close()

    // Execute the query
    res, err := db.Exec(sql_string)
    if err != nil {
        panic(err.Error()) 
    }

    log.Println("Exec Result", res)
}

func GetContent(query string, engine_name string, url string)(content string){
    log.Println("-----------GetContent Start------------")
    log.Println("query = ",query)
    log.Println("engine_name = ",engine_name)
    log.Println("url = ",url)

    c := &soso_proto.SosoCrawler{}
    c.Ua = proto.String(USER_AGENT)
    c.Query = proto.String(query)
    c.Engine = proto.String(engine_name)
    c.Url = proto.String(url)

    raddr,e := net.ResolveTCPAddr("tcp", *server_ip)
    if e != nil {
        log.Println("Resolev TCP error. ", e)
        os.Exit(1)
    }
    con, err := net.DialTCP("tcp", nil, raddr)
    defer con.Close()
    ssn_utils.CheckError(err, "connection error Host not found")

    w_buf, w_err := proto.Marshal(c)
    ssn_utils.CheckError(w_err, "Error Marshal")

    in, err := con.Write(w_buf)
    ssn_utils.CheckError(err, "Error sending data")
    log.Println("in = " , in)

    err = con.CloseWrite()
    ssn_utils.CheckError(err, " con CloseWrite")


    buf := make([]byte, 800000)
    totalBytes := 0
    var rec_buf []byte 
    for {
        n, err := con.Read(buf)
        totalBytes += n
        // Was there an error in reading ?
        if err != nil {
            if err != io.EOF {
                log.Printf("Read error: %s", err)
            }
            break
        }
        rec_buf = append(rec_buf, buf[0:n] ...) 
    }

    res := &soso_proto.SosoCrawlerResp{}
    err = proto.Unmarshal(rec_buf, res)

    log.Println("-----------GetContent End------------")
    return res.GetContent()
}


func ParseHtml(engine_name string, html_string string) (res map[int]string){
    log.Println("-----------Pase Html Start------------")
    log.Println("engine_name = ", engine_name)

    time_now := time.Now()
    log.Println("parse_server_ip = ", *parse_server_ip)

    raddr,e := net.ResolveTCPAddr("tcp", *parse_server_ip)
    ssn_utils.CheckError(e, "ResolveTCPAddr")

    client, err := net.DialTCP("tcp", nil, raddr)
    defer client.Close()
    ssn_utils.CheckError(err, "dialing...")

    err = client.SetKeepAlive(true)
    ssn_utils.CheckError(err, "TCP SetKeepAlive") 
    //err = client.SetReadBuffer(512)
    //ssn_utils.CheckError(err, "TCP SetReadBuffer")


    //set req proto
    req := &rpc_proto.SearchResultPageRequest{}
    req.Html = []byte(html_string)
    req.Engine = []byte(engine_name)
    req.IsRebuild = proto.Bool(true)
    w_buf, w_err := proto.Marshal(req)
    ssn_utils.CheckError(w_err, " req proto ->[]byte ")

    //set rpc proto here
    rpc_p := &rpc_proto.Request{}
    rpc_p.ServiceName = proto.String("evasys.SrpprsService")
    rpc_p.MethodName = proto.String("Parse")
    rpc_p.RequestProto = w_buf

    ww_buf, ww_err := proto.Marshal(rpc_p)
    ssn_utils.CheckError(ww_err, " rpc proto -> []byte")

    in, err := client.Write(ww_buf)
    ssn_utils.CheckError(err, " client write ")
    log.Println("in = " , in)

    err = client.CloseWrite()
    ssn_utils.CheckError(err, " client CloseWrite")


    var rec_buf []byte
    buf := make([]byte, 800000)
    t_len := 0
    for {
        n, err := client.Read(buf)
        // Was there an error in reading ?
        if err != nil {
            if err != io.EOF {
                log.Printf("Read error: %s", err)
            }
            break
        }
        t_len += n
        rec_buf = append(rec_buf, buf[0:n] ...)
    }

    log.Println("t_len", t_len)
    log.Println("buf_len", len(rec_buf))

    rpc_response :=  &rpc_proto.Response{}
    err = proto.Unmarshal(rec_buf, rpc_response)
    ssn_utils.CheckError(err, "unmarshal rpc response err")

    result_buf := rpc_response.GetResponseProto()

    result_response := &rpc_proto.SearchResultPageResponse{}
    err = proto.Unmarshal(result_buf, result_response)

    log.Println("---------------------------")
    //log.Println(rec_buf)
    log.Println(err)

    log.Println("---------------------------")
    parse_results := result_response.GetSearchResult()

    result := make(map[int]string)
    for _, item := range parse_results {
        result[int(item.GetRank())] = string(item.GetUrl())
    }
    log.Println("time = ", time.Since(time_now))

    log.Println("-----------Pase Html End------------")
    return result 
}
