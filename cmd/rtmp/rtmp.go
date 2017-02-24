// Created on 2016-04-05
// summary: ScreenStreamer
// author: YangHaitao

package main

import (
	"flag"
	"fmt"
	"go-gypsy/yaml"
	"image"
	"log"
	"strings"
	// "image/jpeg"

	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"ScreenStreamer/config"
	"ScreenStreamer/lib/src/codec"
	logger "ScreenStreamer/logger_seelog"
	_ "net/http/pprof"
	"screenshot"
	"seelog"
	"simplertmp/rtmp"

	"flv.go/flv"
)

var Log seelog.LoggerInterface
var Config *yaml.File
var ServerHost string
var ServerPort string
var Threads int
var Fps int
var BitRate int
var Mode string
var Buffer_Queue_Size int
var Alpha int
var Done bool = false
var ToSBS bool = false
var Cursor bool = false

func GetFirstAvc(c *codec.H264Encoder, width, height uint16) *flv.AVCVideoFrame {
	avc := new(flv.AVCVideoFrame)
	avc.VideoFrame = new(flv.VideoFrame)
	avc.VideoFrame.CFrame = new(flv.CFrame)
	avc.Stream = 0
	avc.Dts = uint32(0)
	avc.Type = flv.TAG_TYPE_VIDEO
	avc.Flavor = flv.KEYFRAME
	avc.CodecId = flv.VIDEO_CODEC_AVC
	avc.Width = width
	avc.Height = height
	avc.PacketType = flv.VIDEO_AVC_SEQUENCE_HEADER
	avc.Body = make([]byte, 0)
	avc.Body = append(avc.Body, (uint8(flv.VIDEO_FRAME_TYPE_KEYFRAME)<<4)|uint8(avc.CodecId))
	avc.Body = append(avc.Body, uint8(flv.VIDEO_AVC_SEQUENCE_HEADER))

	SPS := make([]byte, 0)
	PPS := make([]byte, 0)
	headerS := fmt.Sprintf("%X", c.Header)
	i := strings.Index(headerS, "0000000167")
	if i != -1 {
		spsTmp := c.Header[i/2+4:]
		spsTmpS := fmt.Sprintf("%X", spsTmp)
		j := strings.Index(spsTmpS, "0000000168")
		if j != -1 {
			sps := spsTmp[:j/2]
			pps := spsTmp[j/2+4:]
			SPS = append(SPS, sps...)
			PPS = append(PPS, pps...)
		}
	}

	avc.Body = append(avc.Body, 0)
	avc.Body = append(avc.Body, 0)
	avc.Body = append(avc.Body, 0)

	avc.Body = append(avc.Body, 0x01)
	avc.Body = append(avc.Body, 0x64)
	avc.Body = append(avc.Body, 0x00)
	avc.Body = append(avc.Body, 0x0A)
	avc.Body = append(avc.Body, 0xFF) // reserved | NALULengthSizeMinusOne
	avc.Body = append(avc.Body, 0xE1) // reserved | number of sps NALUs
	lenSPS := len(SPS)
	lenPPS := len(PPS)
	avc.Body = append(avc.Body, uint8(lenSPS>>8)&0xFF)
	avc.Body = append(avc.Body, uint8(lenSPS&0xFF))
	avc.Body = append(avc.Body, SPS...)
	avc.Body = append(avc.Body, 0x01)
	avc.Body = append(avc.Body, uint8(lenPPS>>8&0xFF))
	avc.Body = append(avc.Body, uint8(lenPPS&0xFF))
	avc.Body = append(avc.Body, PPS...)
	return avc
}

func CaptureScreenMustAvc(c *codec.H264Encoder, dts uint32) *flv.AVCVideoFrame {
	t := time.Now()
	avc := new(flv.AVCVideoFrame)
	avc.VideoFrame = new(flv.VideoFrame)
	avc.VideoFrame.CFrame = new(flv.CFrame)
	img := screenshot.CaptureScreenYCbCrMust(false, false)
	tt := time.Now()
	data, err := c.Encode(img)
	if err != nil {
		data, err = c.Encode(img)
	}
	for len(data.Data) == 0 {
		data, err = c.Encode(img)
	}
	ttt := time.Now()
	avc.Stream = 0
	avc.Dts = dts
	avc.Type = flv.TAG_TYPE_VIDEO
	avc.Flavor = flv.KEYFRAME
	avc.CodecId = flv.VIDEO_CODEC_AVC
	avc.Width = uint16(img.Rect.Dx())
	avc.Height = uint16(img.Rect.Dy())
	avc.PacketType = flv.VIDEO_AVC_NALU
	avc.Body = make([]byte, 0)
	if data.Key {
		avc.Body = append(avc.Body, (uint8(flv.VIDEO_FRAME_TYPE_KEYFRAME)<<4)|uint8(avc.CodecId))
	} else {
		avc.Body = append(avc.Body, (uint8(flv.VIDEO_FRAME_TYPE_INTER_FRAME)<<4)|uint8(avc.CodecId))
	}
	avc.Body = append(avc.Body, uint8(flv.VIDEO_AVC_NALU))
	avc.Body = append(avc.Body, 0)
	avc.Body = append(avc.Body, 0)
	avc.Body = append(avc.Body, 0)
	if data.Data[0] == 0 && data.Data[1] == 0 && data.Data[2] == 0 && data.Data[3] == 1 {
		lenBody := len(data.Data[4:])
		avc.Body = append(avc.Body, uint8(lenBody>>24)&0xFF)
		avc.Body = append(avc.Body, uint8(lenBody>>16)&0xFF)
		avc.Body = append(avc.Body, uint8(lenBody>>8)&0xFF)
		avc.Body = append(avc.Body, uint8(lenBody&0xFF))
		avc.Body = append(avc.Body, data.Data[4:]...)
	} else if data.Data[0] == 0 && data.Data[1] == 0 && data.Data[2] == 1 {
		lenBody := len(data.Data[3:])
		avc.Body = append(avc.Body, uint8(lenBody>>24)&0xFF)
		avc.Body = append(avc.Body, uint8(lenBody>>16)&0xFF)
		avc.Body = append(avc.Body, uint8(lenBody>>8)&0xFF)
		avc.Body = append(avc.Body, uint8(lenBody&0xFF))
		avc.Body = append(avc.Body, data.Data[3:]...)
	}
	tttt := time.Now()
	log.Println(fmt.Sprintf("Capture: %v, Encode: %v, Build: %v", tt.Sub(t), ttt.Sub(tt), tttt.Sub(ttt)))
	return avc
}

func init() {
	configPath := "./configuration.rtmp.yml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Init Config (%s) error: (%s) does not exist!\n", configPath)
		os.Exit(1)
	} else {
		Config = config.GetConfig(configPath)
	}

	logPath, err := Config.Get("log_path")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['log_path'] error: %s\n", err))
		os.Exit(1)
	}

	logLevel, err := Config.Get("log_level")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['log_level'] error: %s\n", err))
		os.Exit(1)
	}
	// fmt.Printf("main init logger start\n")
	// Log = logger.NewLogger(logPath, logLevel)
	// 20M = 20971520
	Logger, err := logger.NewLogger("main", logPath, "rtmp.log", logLevel, "size", "20971520", "5", true)
	if err != nil {
		fmt.Printf(fmt.Sprintf("Init logger error: %s\n", err))
		os.Exit(1)
	}
	Log = *Logger

	tmp_server_host, err := Config.Get("server_host")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['server_host'] error: %s\n", err))
		os.Exit(1)
	}
	ServerHost = tmp_server_host

	tmp_server_port, err := Config.Get("server_port")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['server_host'] error: %s\n", err))
		os.Exit(1)
	}
	ServerPort = string(tmp_server_port) // strconv.FormatInt(tmp_server_port, 10)

	buffer_queue_size_tmp, err := Config.GetInt("buffer_queue_size")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['buffer_queue_size'] error: %s\n", err))
		os.Exit(1)
	}
	Buffer_Queue_Size = int(buffer_queue_size_tmp)

	fps_tmp, err := Config.GetInt("fps")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['fps'] error: %s\n", err))
		os.Exit(1)
	}
	Fps = int(fps_tmp)

	bit_rate_tmp, err := Config.GetInt("bit_rate")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['bit_rate'] error: %s\n", err))
		os.Exit(1)
	}
	BitRate = int(bit_rate_tmp)

	rtmp.InitBuf(Buffer_Queue_Size)

	to_sbs_tmp, err := Config.GetBool("to_sbs")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['to_sbs'] error: %s\n", err))
		os.Exit(1)
	}
	ToSBS = to_sbs_tmp

	cursor_tmp, err := Config.GetBool("cursor")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['cursor'] error: %s\n", err))
		os.Exit(1)
	}
	Cursor = cursor_tmp

	alpha_tmp, err := Config.GetInt("alpha")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['alpha'] error: %s\n", err))
		os.Exit(1)
	}
	Alpha = int(alpha_tmp)

	rtmp.InitCap(Fps, Alpha)

	mode_tmp, err := Config.Get("mode")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['mode'] error: %s\n", err))
		os.Exit(1)
	}
	Mode = mode_tmp

	threads_tmp, err := Config.GetInt("threads")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['threads'] error: %s\n", err))
		os.Exit(1)
	}
	Threads = int(threads_tmp)
	threads := runtime.GOMAXPROCS(Threads)
	Log.Info(fmt.Sprintf("Server with threads: %d", threads))
	Log.Info(fmt.Sprintf("Server config path: %s", configPath))
	Log.Info(fmt.Sprintf("Server log path: %s", logPath))
	Log.Info(fmt.Sprintf("Server log level: %s", logLevel))
	Log.Info(fmt.Sprintf("Server host: %s", ServerHost))
	Log.Info(fmt.Sprintf("Server port: %s", ServerPort))
}

func worker(c *codec.H264Encoder, img *image.YCbCr) {
	std_interval := float64(1.0 / float64(Fps))
	time_to_sleep := std_interval
	s := time.Now()
	dts := uint32(0)
	avc := GetFirstAvc(c, uint16(img.Rect.Dx()), uint16(img.Rect.Dy()))
	rtmp.Buffer <- avc
	for {
		if Done {
			return
		}
		if dts == 0xffffff {
			dts = uint32(0)
		}
		t := time.Now()
		avc := CaptureScreenMustAvc(c, dts)
		tt := time.Now()
		log.Println(fmt.Sprintf("CaptureScreenMustAvc use: %v", tt.Sub(t)))
		rtmp.Buffer <- avc
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
		Log.Debug(fmt.Sprintf("sleep: %v", sleep_time))
		dts += uint32(std_interval * 1000)
	}
}

func main() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	screenshot.InitConn()
	img := screenshot.CaptureScreenYCbCrMust(false, false)
	c, err := codec.NewH264Encoder(img.Rect.Dx(), img.Rect.Dy(), 0, Fps, 1, Fps, BitRate, image.YCbCrSubsampleRatio444, "bufsize,0k,0", "pixel_format,yuv444p,0")
	if err != nil {
		panic(err)
	}

	var swg sync.WaitGroup

	flag.Parse()
	err = rtmp.ListenAndServe(fmt.Sprintf("%v:%v", ServerHost, ServerPort))
	if err != nil {
		panic(err)
	}
	log.Printf("Rtmp Server Listen At %v:%v", ServerHost, ServerPort)
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	if Mode == "single" {
		go worker(c, img)
	} else {
		go worker(c, img)
	}

	select {
	case signal := <-sigs:
		Log.Info(fmt.Sprintf("Got signal: (%v)", signal))
	}

	defer Log.Flush()
	screenshot.CloseConn()
	Done = true
	Log.Info(fmt.Sprintf("Stopping server"))
	Log.Info(fmt.Sprintf("Server will stop in 5 seconds ..."))
	time.Sleep(5)
	Log.Info(fmt.Sprintf("Waiting on server"))
	swg.Wait()
	Log.Info("Server exit!")
	os.Exit(0)
}
