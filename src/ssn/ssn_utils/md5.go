//package main 
package ssn_utils 

import(
    "log"
    "crypto/md5"
    "fmt"
)

func GetMD5(input string)  (md5_str string) {
    m := md5.New() 
    m.Write([]byte(input))
    md5_str = fmt.Sprintf("%x", m.Sum(nil))
    
    log.Println("input=#" + input + "#, md5=" + md5_str)

    return md5_str
}

/*
func main(){
    input := "hello world"

    h := md5.New()
    h.Write([]byte(input))
    hh := fmt.Sprintf("%x", h.Sum(nil))

    hhh := GetMD5(input)
    log.Printf("%x", h.Sum(nil))
    log.Println(hh)
    log.Println(hhh)
}
*/
