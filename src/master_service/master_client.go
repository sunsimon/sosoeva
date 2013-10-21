package main 


import proto "code.google.com/p/goprotobuf/proto"
import (
    "soso_proto"
    "log"
    "time"
    "net"
    "ssn/ssn_utils"
    "io"
    "flag"
    "os"
)

    
func main() {
    server_ip := flag.String("p", "127.0.0.1", "master server ip")
    flag.Parse()

    c := &soso_proto.SosoCrawler{}
    c.Ua = proto.String("hello world")
    c.Query = proto.String("hello world")
    c.Engine = proto.String("soso")
    c.Url = proto.String("http://www.soso.com")


    start_time := time.Now()


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

    log.Println("result :", res.GetContent())
    log.Println("time=", time.Since(start_time))
}
