package ssn_net

import (
    "fmt"
    "strings"
    "net"
    "log"
)

func GetIP(num int) (ip_add string) {

    ifs, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Print(err)
    }

    i := 1
    ip_addr := ""
    for _, v := range ifs {
        //fmt.Println(v)
        ip_addr = strings.SplitN(v.String(), "/", 4)[0]
        log.Println("==================")
        log.Println(ip_addr)
        log.Println("==================")
        if i == num { 
            break
        }
        i ++
    }
    return ip_addr

}
