package main

import proto "code.google.com/p/goprotobuf/proto"
import (
    "soso_proto"
    "database/sql"
    "log"
    "net"
//    "net/url"
    "flag"
//    "os"
    "ssn/ssn_utils"
    "ssn/ssn_net"
    "ssn/ssn_config"
    "io"
    "time"
    _ "mysql"
)
//rpc use 
import "rpc_proto"
//for get query
//import "strings"
import "strconv"

//Global Var
var parse_server_ip *string
var Config_Reader  *ssn_config.Config 

//db pool
//var db *sql.DB 

const MYSQL_STRING = "jianwensun:sun@tcp(10.11.195.131:3306)/mark_oss"
const MAX_RECEIVE = 1024000

func init(){
    //set flag
    parse_server_ip = flag.String("parse_server_ip", "10.136.30.20:30026", "parse the html page  server ip")
    flag.PrintDefaults()
    flag.Parse()
    
    Config_Reader = ssn_config.Register("vr_monitor.conf")
    config_reader_err := Config_Reader.Read()
    log.Println(config_reader_err)

}

func main(){
    service_port := Config_Reader.Get("SERVER", "port")
    log.Println("service port :", service_port)
    ip_interface_string:= Config_Reader.Get("SERVER", "ip_interface")
    log.Println("ip interface :", ip_interface_string)
    ip_interface_num, _ := strconv.Atoi(ip_interface_string)

    service := ssn_net.GetIP(ip_interface_num) + ":" + service_port
    log.Println("service ip:", service)

    listener, err := net.Listen("tcp", service)
    ssn_utils.CheckError(err, "service listen")

    log.Print("server: listening")

    for i := 1; ; {
        conn, err := listener.Accept()
        ssn_utils.CheckError(err, "server: accept")
        log.Printf("server: accepted from %s", conn.RemoteAddr())
        i ++
        go HandleClient(conn, i)
    }
}

func HandleClient(conn net.Conn,  counter int) {
    defer conn.Close()

    buf := make([]byte, MAX_RECEIVE)
    for {
        log.Print("server: conn: waiting")
        n, err := conn.Read(buf)
        if err != nil {
            if err != io.EOF {
                log.Printf("server: conn: read: %s", err)
            }
            break
        }
        log.Printf("[DEBUG]server input:", buf[:n])


        sogou_meta := &soso_proto.SogouMetaSend{}
        err = proto.Unmarshal(buf[:n], sogou_meta)

        if err != nil{
            log.Println("unSerialize error")
            break
        }
        log.Println("unSerialize :", sogou_meta)
        log.Println("query :", sogou_meta.GetQuery())

        //compose mm
        //lazy to rewrite server 
        content := "hello world"
        res := &soso_proto.SogouMetaResp{}
        res.Status = proto.Int32(1)
        res.Content= proto.String(content)
        res_buf, _ := proto.Marshal(res)

        n, err = conn.Write(res_buf)
        log.Printf("server: conn: wrote %d bytes", n)

        if err != nil {
            log.Printf("server: write: %s", err)
        }
        break
    }
    log.Println("server: conn: closed")
}


func GetResultSQL(sql_string string) (res map[int]map[string]string) {
    log.Println("sql:", sql_string)

    // Open database connection
    db, err := sql.Open("mysql", MYSQL_STRING)
    if err != nil {
        panic(err.Error())  
    }
    defer db.Close()
    db.SetMaxIdleConns(100)

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
    db, err := sql.Open("mysql", MYSQL_STRING)
    if err != nil {
        panic(err.Error())  
    }
    defer db.Close()
    db.SetMaxIdleConns(100)

    // Execute the query
    res, err := db.Exec(sql_string)
    if err != nil {
        panic(err.Error()) 
    }

    log.Println("Exec Result", res)
}

//TODO
func GetContent()(content string){
    return "" 
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

func InsertSnapShortSQL(query, url, html_string string, engine_id, is_search_page int) {

    // Open database connection
    db, err := sql.Open("mysql", MYSQL_STRING)
    if err != nil {
        panic(err.Error())  
    }
    defer db.Close()
    db.SetMaxIdleConns(100)
    // Prepare statement for inserting data
    stmtIns, err := db.Prepare("INSERT INTO tb_dcg_snapshot VALUES( null, ?, ?, ?, ?, ? )") // ? = placeholder
    if err != nil {
        panic(err.Error()) 
    }
    defer stmtIns.Close() // Close the statement when we leave 

    _, err = stmtIns.Exec(query, url, engine_id, html_string, is_search_page) // Insert 
    if err != nil {
        panic(err.Error())
    }

}


