package main

import "os"
import "fmt"
import "strings"
import    _ "mysql"
import "database/sql"
import "encoding/json"
import "net/url"

const MYSQL_STRING = "jianwensun:sun@tcp(10.11.195.131:3306)/mark_oss"

type Response struct {
    RetTid string
    Retcode string
    Retstring string
    data string
    RetSql string
}

func main() {
    res := &Response{}
    res.Retcode = "-1"

    Print_header()
    content := os.Getenv("QUERY_STRING")

    pairs := strings.Split(content, "&")

    params := make(map[string]string)
    for _, v1 := range pairs {
        pair := strings.Split(v1, "=")
        if len(pair) > 0 {
            params[pair[0]] = pair[1]
        }
    }
    var tid, page_type, query, engine_id string
    var is_exist bool

    if tid, is_exist = params["tid"]; !is_exist {
        res.Retstring = "no tid found"
        Print_Res(res)
    }

    res.RetTid = tid

    if page_type, is_exist = params["page_type"]; !is_exist {
        res.Retstring = "no type found"
        Print_Res(res)
    }

    if query, is_exist = params["query"]; !is_exist {
        res.Retstring = "no query word found"
        Print_Res(res)
    }
    query, _ = url.QueryUnescape(query)

    if engine_id, is_exist = params["engine_id"]; !is_exist {
        res.Retstring = "no engine_id found"
        Print_Res(res)
    }

    var sql_string string
    sql_string = "SELECT query, url, html_page from tb_dcg_snapshot "
    if page_type == "page_search" {
        sql_string += "where page_type=0 and query='" + query + "'"
        sql_string += " and engine_id=" + engine_id
        sql_string += " and task_id=" + tid 
        sql_string  += " limit 1"

        temp_res := GetResultSQL(sql_string)

        if len(temp_res) < 1 {
            res.RetSql = sql_string
            res.Retstring = " find no result "
            Print_Res(res)
        }

        for _, item := range temp_res {
            fmt.Println(item["html_page"])
        }
    } else if page_type == "page_result" {
        var rank string
        if rank, is_exist = params["rank"]; !is_exist {
            res.Retstring = "no input rank found"
            Print_Res(res)
        }

        sql_string  += " where page_type=1 and query='" + query + "'"
        sql_string  += " and rank='" + rank + "'"
        sql_string  += " and engine_id=" + engine_id
        sql_string += " and task_id=" + tid 
        sql_string  += " limit 1"

        temp_res := GetResultSQL(sql_string)

        if len(temp_res) < 1 {
            res.RetSql = sql_string
            res.Retstring = " find no result "
            Print_Res(res)
        }

        for _, item := range temp_res {
            fmt.Println(item["html_page"])
        }


    } 

    res.Retstring = "input page_type is invalid"
    Print_Res(res)

    os.Exit(1)

}

func Print_header(){
    fmt.Println("Content-type: text/html\n\n")

}


func Print_Res(v interface{}){
    json_string, _ := json.Marshal(v)
    fmt.Println(string(json_string))
    os.Exit(1)
}

func GetResultSQL(sql_string string) (res map[int]map[string]string) {
    // Open database connection
    db, err := sql.Open("mysql", MYSQL_STRING)
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
            fmt.Println("in rows.Scan")
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
            temp_store[columns[i]] = value
        }
        result[count] = temp_store
        count ++
        fmt.Println("SnapShot Here <BR/>")
        fmt.Println("-----------------------------------")
    }
    return result 
}



