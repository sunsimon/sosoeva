package common
import(
    "time"
    "fmt"
    "net/http"
    //"io/ioutil"
)

func Timeecho() {
    start := time.Now()
    http.Get("http://wap.soso.com")
    fmt.Println(time.Since(start))
}

/*
func main() {
    start := time.Now()

    http.Get("http://wap.soso.com")
    fmt.Println(time.Since(start))
}
*/

