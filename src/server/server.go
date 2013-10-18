package main

import (
    "io"
    "log"
    "net"
    "encoding/json"
    "common"
    "ssn/ssn_net"
    "ssn/ssn_utils"
    "ssn/ssn_config"
    "strconv"
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

func handleClient(conn net.Conn, query string) {
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

        var mm common.Message
        res := json.Unmarshal(buf[:n], &mm)
        if res != nil {
            panic("json decode error")
        }
        log.Println(mm)

        //query_urlencode := ssn_utils.UrlEncode(string(buf[:n]))
        //content := common.GetHttpContent(mm)
        content := mm.GetHttpContent()
        //log.Println(content)


        n, err = conn.Write([]byte(content))
        log.Printf("server: conn: wrote %d bytes", n)

        if err != nil {
            log.Printf("server: write: %s", err)
        }
        break
    }
    log.Println("server: conn: closed")
}
