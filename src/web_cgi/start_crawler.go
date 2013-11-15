package main

import "os"
import "os/exec"
import "fmt"
import "strings"

func main() {
    Print_header()
    content := os.Getenv("QUERY_STRING")
    fmt.Println("input is :", content)

    pairs := strings.Split(content, "&")

    params := make(map[string]string)
    for _, v1 := range pairs {
        pair := strings.Split(v1, "=")
        if len(pair) > 0 {
            params[pair[0]] = pair[1]
        }
    }

    if resume, is_exist := params["resume"]; is_exist {
        fmt.Println("resume=", resume)
        Run(resume)
        os.Exit(1)
    }

    if tid, is_exist := params["tid"]; is_exist {
        fmt.Println("tid=", tid)
        Run(tid)
        os.Exit(1)
    }
}

func Print_header(){
    fmt.Println("Content-type: text/html\n\n")

}

func Run(tid string){
    fmt.Println("start to run", "tid=", tid)
    cmd := exec.Command("/usr/local/apache/cgi-bin/oss_work/Start.sh", tid)
    cmd.Stdin = strings.NewReader("some input")
    err := cmd.Run()
    if err != nil {
        fmt.Println(err)
    }
}

