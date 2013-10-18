package ssn_utils

import (
    "log"
)

func CheckError(err error, info string) {
    if err != nil {
        log.Fatal(info, ". ", err.Error())
    }
}
