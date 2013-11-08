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
    _ "mysql"
)

var server_ip *string

const USER_AGENT = "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/32.0.1664.3"

func init(){
        server_ip = flag.String("server_ip", "127.0.0.1", "master server ip")
        flag.Parse()
}

func main() {
    // Open database connection
    db, err := sql.Open("mysql", "root:sun@/debt")
    if err != nil {
        panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
    }
    defer db.Close()

    // Execute the query
    rows, err := db.Query("SELECT creditor_id, creditor_name FROM debt")
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

    // Fetch rows
    for rows.Next() {
        // get RawBytes from data
        err = rows.Scan(scanArgs...)
        if err != nil {
            panic(err.Error()) // proper error handling instead of panic in your app
        }

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
        }
        log.Println("-----------------------------------")
    }
}

func GetContent(query string)(content string){
    log.Println("query = ",query)

    c := &soso_proto.SosoCrawler{}
    c.Ua = proto.String("hello world")
    c.Query = proto.String("hello world")
    c.Engine = proto.String("soso")
    c.Url = proto.String("http://www.soso.com")

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

    buf := make([]byte, 800000)
    totalBytes := 0
    html_string := ""
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
        html_string += string(buf[0:n])
    }

    res := &soso_proto.SosoCrawlerResp{}
    err = proto.Unmarshal([]byte(html_string), res)

    return res.GetContent()
}
