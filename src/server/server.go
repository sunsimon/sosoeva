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
    "sync"
    "strings"
    "time"
)

const MAX_RECEIVE = 1024000
var Time_Recorder *TimeRecord

type Record struct {
    Engine string
    Latest_Time int
    Time_Gap int
    Locker *sync.Mutex
}

type TimeRecord struct {
    SOSO *Record
    BAIDU *Record
    SOGOU *Record

    BING *Record
    QIHU *Record
    GOOGLE *Record

    SOGOU_TEST *Record
}

func RegisterRecord(engine string, time_gap int) *Record {
    engine = strings.ToLower(engine)
    log.Println("time record regist. (", engine, time_gap, ")")
    return &Record{engine, 0, time_gap, new(sync.Mutex)}
}

func (self *TimeRecord) GetLatestTime(engine string)(latest_time int, time_gap int) {
    engine = strings.ToLower(engine) 
    switch engine {
    case "soso" :
        self.SOSO.Locker.Lock()
        defer self.SOSO.Locker.Unlock()
        return self.SOSO.Latest_Time, self.SOSO.Time_Gap
    case "baidu" :
        self.BAIDU.Locker.Lock()
        defer self.BAIDU.Locker.Unlock()
        return self.BAIDU.Latest_Time, self.BAIDU.Time_Gap
    case "sogou" :
        self.SOGOU.Locker.Lock()
        defer self.SOGOU.Locker.Unlock()
        return self.SOGOU.Latest_Time, self.SOGOU.Time_Gap
    case "google" :
        self.GOOGLE.Locker.Lock()
        defer self.GOOGLE.Locker.Unlock()
        return self.GOOGLE.Latest_Time, self.GOOGLE.Time_Gap
    case "bing" :
        self.BING.Locker.Lock()
        defer self.BING.Locker.Unlock()
        return self.BING.Latest_Time, self.BING.Time_Gap
    case "360", "360so", "qihu", "qihu360" :
        self.QIHU.Locker.Lock()
        defer self.QIHU.Locker.Unlock()
        return self.QIHU.Latest_Time, self.QIHU.Time_Gap
    }

    return 0, 0
}

func (self *TimeRecord) SetLatestTime(engine string){
    engine = strings.ToLower(engine) 
    switch engine {
    case "soso" :
        self.SOSO.Locker.Lock()
        defer self.SOSO.Locker.Unlock()
        self.SOSO.Latest_Time = time.Now().Nanosecond()
    case "baidu" :
        self.BAIDU.Locker.Lock()
        defer self.BAIDU.Locker.Unlock()
        self.BAIDU.Latest_Time = time.Now().Nanosecond()
    case "sogou" :
        self.SOGOU.Locker.Lock()
        defer self.SOGOU.Locker.Unlock()
        self.SOGOU.Latest_Time = time.Now().Nanosecond()
    case "google" :
        self.GOOGLE.Locker.Lock()
        defer self.GOOGLE.Locker.Unlock()
        self.GOOGLE.Latest_Time = time.Now().Nanosecond()
    case "bing" :
        self.BING.Locker.Lock()
        defer self.BING.Locker.Unlock()
        self.GOOGLE.Latest_Time = time.Now().Nanosecond()
    case "360", "360so", "qihu", "qihu360" :
        self.QIHU.Locker.Lock()
        defer self.QIHU.Locker.Unlock()
        self.QIHU.Latest_Time = time.Now().Nanosecond()
    }

}

func init(){
    //here copy code blow from main
    config_reader := ssn_config.Register("server.conf")
    config_reader_err := config_reader.Read()
    log.Println(config_reader_err)

    soso_time_gap, _ := strconv.Atoi(config_reader.Get("SERVER", "soso_time_gap"))
    baidu_time_gap, _ := strconv.Atoi(config_reader.Get("SERVER", "baidu_time_gap"))
    sogou_time_gap, _ := strconv.Atoi(config_reader.Get("SERVER", "sogou_time_gap"))
    bing_time_gap, _ := strconv.Atoi(config_reader.Get("SERVER", "bing_time_gap"))
    google_time_gap, _ := strconv.Atoi(config_reader.Get("SERVER", "google_time_gap"))
    qihu_time_gap, _ := strconv.Atoi(config_reader.Get("SERVER", "qihu_time_gap"))

    //init time record
    Time_Recorder = &TimeRecord{
        SOSO : RegisterRecord("soso", soso_time_gap),
        BAIDU :RegisterRecord("baidu", baidu_time_gap),
        SOGOU: RegisterRecord("sogou", sogou_time_gap),
        BING : RegisterRecord("bing", bing_time_gap),
        GOOGLE:RegisterRecord("google", google_time_gap),
        QIHU : RegisterRecord("qihu", qihu_time_gap),
    }
}



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

        //check delay
        var content string
        last_time, time_gap := Time_Recorder.GetLatestTime(mm.Engine)
        if time.Now().Nanosecond() - last_time > time_gap {
            //query_urlencode := ssn_utils.UrlEncode(string(buf[:n]))
            content = mm.GetContent()
            //log.Println(content)
        } else {
            content = "000"
        }


        n, err = conn.Write([]byte(content))
        log.Printf("server: conn: wrote %d bytes", n)

        if err != nil {
            log.Printf("server: write: %s", err)
        }
        break
    }
    log.Println("server: conn: closed")
}
