// Created on 2015-08-03
// summary: stop handler
// author: YangHaitao

package handlers

import (
    "fmt"
    "encoding/json"
    "net/http"
)

func StopHandler(w http.ResponseWriter, r *http.Request) {
    headers := make(map[string]string)
    Log.Info(fmt.Sprintf("Stop server"))

    headers[Headers["status"]] = Response["success"]
    headers[Headers["message"]] = Response["server_will_stop"]

    w.Header().Set(Headers["status"], headers[Headers["status"]])
    w.Header().Set(Headers["message"], headers[Headers["message"]])
    w.WriteHeader(http.StatusOK)

    Done <- true

    result, _ := json.Marshal(headers)
    w.Write(result)
}