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
    //"sync"
    //"strings"
    "math"
)

const MAX_RECEIVE = 1024000
const MAX_SERVER_NUM = 10


var Server_List = []string{}
var Config_Reader  *ssn_config.Config

func init(){
    Config_Reader = ssn_config.Register("server.conf")

    config_reader_err := Config_Reader.Read()
    log.Println(config_reader_err)

    //server ip list
    for i := 1; i < MAX_SERVER_NUM; i ++ {
        server_config_string := "server_" + strconv.Itoa(i) 
        server_ip :=  Config_Reader.Get("MASTER_SERVER", server_config_string)
        if server_ip != "" {
            log.Println("server_ip", server_ip)
            if checkConnection(server_ip) {
                Server_List = append(Server_List, server_ip)
            } else {
                log.Println("server_ip", server_ip, "not active")
            }
        }
    }
    log.Println("Server_List", Server_List)

}

func main() {
    service_port := Config_Reader.Get("MASTER_SERVER", "port")
    log.Println("service port :", service_port)
    ip_interface_string:= Config_Reader.Get("MASTER_SERVER", "ip_interface")
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
        mm.Query = soso_crawler.GetQuery()
        mm.Engine = soso_crawler.GetEngine()
        mm.Url = soso_crawler.GetUrl()
        mm.User_Agent = soso_crawler.GetUa()
        mm.Sleep_Time = 1

        content := deliver(mm, counter)

        res := &soso_proto.SosoCrawlerResp{}
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

func deliver(mm *common.Message, counter int) (res string) {

    num := math.Mod(float64(counter), float64(len(Server_List)))
    log.Println("num : ", num)
    download_server_ip := Server_List[int(num)] 

    start_time := time.Now()
    raddr,e := net.ResolveTCPAddr("tcp", download_server_ip)
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

    if len(html_string) < 10 {
        html_string = deliver(mm, counter + 1)
    }
    return html_string
}

func checkConnection(server_ip string) (is_active bool){
    raddr,e := net.ResolveTCPAddr("tcp", server_ip)
    if e != nil {
        log.Println("IP", server_ip, "Resolev TCP error. ", e)
        return false 
    }

    con, err := net.DialTCP("tcp", nil, raddr)
    defer con.Close()
    if err != nil {
        log.Println("IP", server_ip, "Dial err", err)
        return false
    }

    //todo
    //try get one page here

    return true
}
