// Created on 2015-07-21
// summary: put handler
// author: YangHaitao

package handlers

import (
	"fmt"
	"image"
	logger "ScreenStreamer/logger_seelog"
	"seelog"
	"stringio"
)

var Log seelog.LoggerInterface
var Headers map[string]string
var Response map[string]string
var maxMemory int64 = 1024 * 1024
var Done chan bool
var Buffer chan *image.RGBA
var BufferArray [10]chan *image.RGBA
var Images chan *stringio.StringIO
var ImageBuffer *stringio.StringIO
var ImagesArray [10]chan *stringio.StringIO
var TasksArray [10]chan int
var Fps int
var Alpha int
var Broadcast bool

var Quality int
var Left int
var Top int
var Width int
var Height int

func InitLog() {
	Logger, err := logger.GetLogger("main")
	if err != nil {
		fmt.Printf("GetLogger error: %s\n", err)
	}
	fmt.Printf("GetLogger: %s\n", "init")
	Log = *Logger
}

func GetImage() *stringio.StringIO {
	select {
    case snapshot := <-Images:
        ImageBuffer = snapshot
		return ImageBuffer
    default:
		return ImageBuffer
    }
}

func InitBuf(b_size, i_size, cb_size, ci_size, ct_size int) {
	Buffer = make(chan *image.RGBA, b_size)
	BufferArray[0] = make(chan *image.RGBA, cb_size)
	BufferArray[1] = make(chan *image.RGBA, cb_size)
	BufferArray[2] = make(chan *image.RGBA, cb_size)
	BufferArray[3] = make(chan *image.RGBA, cb_size)
	BufferArray[4] = make(chan *image.RGBA, cb_size)
	BufferArray[5] = make(chan *image.RGBA, cb_size)
	BufferArray[6] = make(chan *image.RGBA, cb_size)
	BufferArray[7] = make(chan *image.RGBA, cb_size)
	BufferArray[8] = make(chan *image.RGBA, cb_size)
	BufferArray[9] = make(chan *image.RGBA, cb_size)
	Images = make(chan *stringio.StringIO, i_size)
	ImagesArray[0] = make(chan *stringio.StringIO, ci_size)
	ImagesArray[1] = make(chan *stringio.StringIO, ci_size)
	ImagesArray[2] = make(chan *stringio.StringIO, ci_size)
	ImagesArray[3] = make(chan *stringio.StringIO, ci_size)
	ImagesArray[4] = make(chan *stringio.StringIO, ci_size)
	ImagesArray[5] = make(chan *stringio.StringIO, ci_size)
	ImagesArray[6] = make(chan *stringio.StringIO, ci_size)
	ImagesArray[7] = make(chan *stringio.StringIO, ci_size)
	ImagesArray[8] = make(chan *stringio.StringIO, ci_size)
	ImagesArray[9] = make(chan *stringio.StringIO, ci_size)
	TasksArray[0] = make(chan int, ct_size)
	TasksArray[1] = make(chan int, ct_size)
	TasksArray[2] = make(chan int, ct_size)
	TasksArray[3] = make(chan int, ct_size)
	TasksArray[4] = make(chan int, ct_size)
	TasksArray[5] = make(chan int, ct_size)
	TasksArray[6] = make(chan int, ct_size)
	TasksArray[7] = make(chan int, ct_size)
	TasksArray[8] = make(chan int, ct_size)
	TasksArray[9] = make(chan int, ct_size)
}

func InitCap(left, top, width, height, quality, fps, alpha int, broadcast bool) {
	Left = left
	Top = top
	Width = width
	Height = height
	Quality = quality
	Fps = fps
	Alpha = alpha
	Broadcast = broadcast
}

func init() {
	Headers = make(map[string]string)
	Response = make(map[string]string)
	Headers["status"] = "X-Operate-Status"
	Headers["message"] = "X-Operate-Message"
	Response["success"] = "SUCCESS"
	Response["failure"] = "FAILURE"
	Response["exist"] = "EXIST"
	Response["server_will_stop"] = "Info: streamer server will stop! please use 'ps' or 'curl' to check."
	Done = make(chan bool, 1)
}
