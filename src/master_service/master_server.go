package main

import (
    "io"
    "log"
    "net"
    "ssn/ssn_net"
    "ssn/ssn_utils"
    "ssn/ssn_config"
    "strconv"
    "soso_proto"
    "code.google.com/p/goprotobuf/proto"
    "common"
    "time"
    "encoding/json"
)

const MAX_RECEIVE = 1024000

func main() {

    config_reader := ssn_config.Register("server.conf")
    config_reader_err := config_reader.Read()
    log.Println(config_reader_err)

    service_port := config_reader.Get("SERVER", "port")
    log.Println("service port :", service_port)
    ip_interface_string:= config_reader.Get("SERVER", "ip_interface")
    log.Println("ip interface :", ip_interface_string)
    ip_interface_num, _ := strconv.Atoi(ip_interface_string)

    service := ssn_net.GetIP(ip_interface_num) + ":" + service_port

    log.Println("service ip:", service)

    listener, err := net.Listen("tcp", service)
    ssn_utils.CheckError(err, "service listen")

    log.Print("server: listening")

    for {
        conn, err := listener.Accept()
        ssn_utils.CheckError(err, "server: accept")
        log.Printf("server: accepted from %s", conn.RemoteAddr())

        go handleClient(conn, "0")
    }
}

func handleClient(conn net.Conn, optional string) {
    defer conn.Close()

    log.Println("optional = ", optional)

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

 
        soso_crawler := &soso_proto.SosoCrawler{}
        err = proto.Unmarshal(buf[:n], soso_crawler)

        if err != nil{
            log.Println("unSerialize error")
            break
        }
        log.Println("unSerialize :", soso_crawler)
        log.Println("query :", soso_crawler.GetQuery())

        //compose mm
        //lazy to rewrite server 
        mm := new(common.Message)
        mm.Query = "hello"
        mm.Engine = "soso"
        mm.Url = "http://www.soso.com/q?pid=s.idx&cid=s.idx.se&w="
        mm.User_Agent = "hello world"
        mm.Sleep_Time = 1

        
        content := deliver(mm) 

        n, err = conn.Write([]byte(content))
        log.Printf("server: conn: wrote %d bytes", n)

        if err != nil {
            log.Printf("server: write: %s", err)
        }
        break
    }
    log.Println("server: conn: closed")
}

func deliver(mm *common.Message) (res string) {
    server_ip := "218.246.34.209:9999"
    start_time := time.Now()
    raddr,e := net.ResolveTCPAddr("tcp", server_ip)
    if e != nil {
        log.Println("Resolev TCP error. ", e)
    }

    con, err := net.DialTCP("tcp", nil, raddr)
    //con, err := net.Dial("tcp", server_ip)
    defer con.Close()
    ssn_utils.CheckError(err, "connection error Host not found")

    json_data, _ := json.Marshal(mm)

    in, err := con.Write(json_data)
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
    content_len := totalBytes
    log.Println("content_len = ", content_len)

    /*
    query_md5_value := ssn_utils.GetMD5(m.Query)
    filename := m.Engine + "_" + query_md5_value + ".html"
    log.Println(filename)
    doo <- filename

    log.Println("html string len = ", len(html_string))
    common.WriteStringIntoFile(html_string,  filename)
    //common.WriteStringIntoFile(string(buf[:content_len]),  filename)
    //common.WriteStringIntoFile(string(buf[:content_len]),  filename)
    log.Println("query=", m.Query, m.Engine, ". time=", time.Since(start_time))
    */
    log.Println("time=", time.Since(start_time))
    return html_string
}
