package main

import proto "code.google.com/p/goprotobuf/proto"
import "log"
import "rpc_proto"
import "net"
import "io"
import "ssn/ssn_utils"

func main(){

    serverAddress := "10.136.30.20"
    port := "9997"

    server_ip := serverAddress + ":" + port
    log.Println("server_ip = ", server_ip)

    raddr,e := net.ResolveTCPAddr("tcp", server_ip)
    ssn_utils.CheckError(e, "ResolveTCPAddr")

    client, err := net.DialTCP("tcp", nil, raddr)
    defer client.Close()
    ssn_utils.CheckError(err, "dialing...")

    err = client.SetKeepAlive(true)
    ssn_utils.CheckError(err, "TCP SetKeepAlive") 
    err = client.SetReadBuffer(10)
    ssn_utils.CheckError(err, "TCP SetReadBuffer")


    //set req proto
    html_string := "<html><head></head><body></body></html>"
    req := &rpc_proto.SearchResultPageRequest{}
    req.Html = []byte(html_string)
    req.Engine = []byte("baidu")
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
    buf := make([]byte, 65535)
    for {
        n, err := client.Read(buf)
        // Was there an error in reading ?
        if err != nil {
            if err != io.EOF {
                log.Printf("Read error: %s", err)
            }
            break
        }
        rec_buf = append(rec_buf, buf[0:n] ...)
    }

    result_response := &rpc_proto.SearchResultPageResponse{}
    err = proto.Unmarshal(rec_buf, result_response)

    log.Println("---------------------------")
    log.Println(result_response.GetEngine())

}
