// Created on 2015-07-21
// summary: get handler
// author: YangHaitao

package handlers

import (
    "fmt"
    "net/http"
)

func JpegHandler(w http.ResponseWriter, r *http.Request) {
    Log.Debug(fmt.Sprintf("Start request %s", r.URL))

    Log.Debug(fmt.Sprintf("Wait source"))
    snapshot := <- Images
    Log.Debug(fmt.Sprintf("Write snapshot"))

    w.Header().Add("Content-Type", "image/jpeg")
    w.Write(snapshot.GetValueBytes())
    Log.Debug(fmt.Sprintf("Success request"))
}
