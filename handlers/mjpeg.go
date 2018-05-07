// Created on 2015-07-15
// summary: get handler
// author: YangHaitao

package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
	"stringio"
	"html/template"
)

func MjpegHandler(w http.ResponseWriter, r *http.Request) {
	Log.Debug(fmt.Sprintf("Start request %s", r.URL))

	mimeWriter := multipart.NewWriter(w)
	mimeWriter.SetBoundary("--boundary")

	Log.Debug(fmt.Sprintf("Boundary: %s", mimeWriter.Boundary()))

	contentType := fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", mimeWriter.Boundary())
	// contentType := fmt.Sprintf("multipart/x-mixed-replace;boundary=--boundary")
	w.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate, pre-check=0, post-check=0, max-age=0")
	w.Header().Add("Content-Type", contentType)
	w.Header().Add("Pragma", "no-cache")
	w.Header().Add("Connection", "close")
	// w.Header().Add("X-Framerate", "30")

	n := 0
	s := time.Now()
	for {
		if n == 30 {
			ss := time.Now()
			fps := float64(30) / ss.Sub(s).Seconds()
			Log.Info(fmt.Sprintf("%.2ffps", fps))
			n = 0
			s = time.Now()
		}

		partHeader := make(textproto.MIMEHeader)
		partHeader.Add("Content-Type", "image/jpeg")
		partHeader.Add("X-StartTime", fmt.Sprintf("%v", s.Unix()))
		partHeader.Add("X-Timestamp", fmt.Sprintf("%v", s.Unix()))

		partWriter, partErr := mimeWriter.CreatePart(partHeader)
		if nil != partErr {
			Log.Error(fmt.Sprintf(partErr.Error()))
			break
		}

		var snapshot *stringio.StringIO

		if Broadcast {
			std_interval := float64(1.0 / float64(Fps))
			time_to_sleep := std_interval
			s := time.Now()
			snapshot = GetImage()
			ss := time.Now()
			interval := ss.Sub(s)
			if interval.Seconds() < std_interval {
				time_to_sleep += float64(Alpha) * time_to_sleep / float64(100)
			} else {
				time_to_sleep -= float64(Alpha) * time_to_sleep / float64(100)
			}
			if time_to_sleep < float64(0) {
				time_to_sleep = float64(0)
			}
			s = ss
			sleep_time, _ := time.ParseDuration(fmt.Sprintf("%fs", time_to_sleep))
			time.Sleep(sleep_time)
		} else {
			snapshot = <-Images
		}

		if _, writeErr := partWriter.Write(snapshot.GetValueBytes()); nil != writeErr {
			Log.Error(fmt.Sprintf(writeErr.Error()))
		}
		n += 1
	}

	mimeWriter.Close()
	Log.Debug(fmt.Sprintf("Success request"))
}

func MjpegViewerHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./templates/mjpeg.html")
	p := struct {
		Test string
	}{
		"this is a test!",
	}
	t.Execute(w, p)
}
