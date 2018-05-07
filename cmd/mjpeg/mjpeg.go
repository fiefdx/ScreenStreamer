// Created on 2016-04-05
// summary: ScreenStreamer
// author: YangHaitao

package main

import (
	"flag"
	"fmt"
	"go-gypsy/yaml"
	"image"
	// "image/jpeg"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	logger "ScreenStreamer/logger_seelog"
	ljpeg "github.com/pixiv/go-libjpeg/jpeg"
	_ "net/http/pprof"
	"screenshot"
	"seelog"
	"stoppableListener"
	"stringio"

	"ScreenStreamer/config"
	"ScreenStreamer/handlers"
)

var Log seelog.LoggerInterface
var Config *yaml.File
var ServerHost string
var ServerPort string
var Threads int
var Fps int
var Quality int
var Left int
var Top int
var Width int
var Height int
var ResizeWidth int
var ResizeHeight int
var Mode string
var Tasks_Queue_Size int
var Buffer_Queue_Size int
var Images_Queue_Size int
var Convert_Buffer_Size int
var Convert_Images_Size int
var Shot int
var Convert int
var Alpha int
var Done bool = false
var ToSBS bool = false
var Broadcast bool = false
var Cursor bool = false

var configPath string

func CaptureWindowMust(pos *screenshot.POS, size *screenshot.SIZE, resize *screenshot.RESIZE, toSBS bool, cursor bool) *image.RGBA {
	img, err := screenshot.CaptureWindow(pos, size, resize, toSBS, cursor)
	errN := 0
	for err != nil {
		Log.Error(fmt.Sprintf("CaptureWindowMust Error: %v", err))
		img, err = screenshot.CaptureWindow(pos, size, resize, toSBS, cursor)
		time.Sleep(time.Duration(10)*time.Millisecond + time.Duration(errN*50)*time.Millisecond)
		if errN < 20 {
			errN += 1
		}
	}
	return img
}

func Init() {
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
	Logger, err := logger.NewLogger("main", logPath, "mjpeg.log", logLevel, "size", "20971520", "5", true)
	if err != nil {
		fmt.Printf(fmt.Sprintf("Init logger error: %s\n", err))
		os.Exit(1)
	}
	Log = *Logger

	handlers.InitLog()

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

	quality_tmp, err := Config.GetInt("quality")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['quality'] error: %s\n", err))
		os.Exit(1)
	}
	Quality = int(quality_tmp)

	tasks_queue_size_tmp, err := Config.GetInt("tasks_queue_size")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['tasks_queue_size'] error: %s\n", err))
		os.Exit(1)
	}
	Tasks_Queue_Size = int(tasks_queue_size_tmp)

	buffer_queue_size_tmp, err := Config.GetInt("buffer_queue_size")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['buffer_queue_size'] error: %s\n", err))
		os.Exit(1)
	}
	Buffer_Queue_Size = int(buffer_queue_size_tmp)

	images_queue_size_tmp, err := Config.GetInt("images_queue_size")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['images_queue_size'] error: %s\n", err))
		os.Exit(1)
	}
	Images_Queue_Size = int(images_queue_size_tmp)

	convert_buffer_size_tmp, err := Config.GetInt("convert_buffer_size")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['convert_buffer_size'] error: %s\n", err))
		os.Exit(1)
	}
	Convert_Buffer_Size = int(convert_buffer_size_tmp)

	convert_images_size_tmp, err := Config.GetInt("convert_images_size")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['convert_images_size'] error: %s\n", err))
		os.Exit(1)
	}
	Convert_Images_Size = int(convert_images_size_tmp)

	fps_tmp, err := Config.GetInt("fps")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['fps'] error: %s\n", err))
		os.Exit(1)
	}
	Fps = int(fps_tmp)

	broadcast_tmp, err := Config.GetBool("broadcast")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['broadcast'] error: %s\n", err))
		os.Exit(1)
	}
	Broadcast = broadcast_tmp

	handlers.InitBuf(Buffer_Queue_Size, Images_Queue_Size, Convert_Buffer_Size, Convert_Images_Size, Tasks_Queue_Size)

	left_tmp, err := Config.GetInt("left")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['left'] error: %s\n", err))
		os.Exit(1)
	}
	Left = int(left_tmp)

	top_tmp, err := Config.GetInt("top")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['top'] error: %s\n", err))
		os.Exit(1)
	}
	Top = int(top_tmp)

	width_tmp, err := Config.GetInt("width")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['width'] error: %s\n", err))
		os.Exit(1)
	}
	Width = int(width_tmp)

	height_tmp, err := Config.GetInt("height")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['height'] error: %s\n", err))
		os.Exit(1)
	}
	Height = int(height_tmp)

	resize_width_tmp, err := Config.GetInt("resize_width")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['resize_width'] error: %s\n", err))
		os.Exit(1)
	}
	ResizeWidth = int(resize_width_tmp)

	resize_height_tmp, err := Config.GetInt("resize_height")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['resize_height'] error: %s\n", err))
		os.Exit(1)
	}
	ResizeHeight = int(resize_height_tmp)

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

	shot_tmp, err := Config.GetInt("shot")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['shot'] error: %s\n", err))
		os.Exit(1)
	}
	if shot_tmp > 10 {
		shot_tmp = 10
	}
	Shot = int(shot_tmp)

	alpha_tmp, err := Config.GetInt("alpha")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['alpha'] error: %s\n", err))
		os.Exit(1)
	}
	Alpha = int(alpha_tmp)

	convert_tmp, err := Config.GetInt("convert")
	if err != nil {
		fmt.Printf(fmt.Sprintf("Get Config['convert'] error: %s\n", err))
		os.Exit(1)
	}
	if convert_tmp > 10 {
		convert_tmp = 10
	}
	Convert = int(convert_tmp)

	handlers.InitCap(Left, Top, Width, Height, Quality, Fps, Alpha, Broadcast)

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

func main() {
	flag.StringVar(&configPath, "config", "./configuration.mjpeg.yml", "-config=./configuration.mjpeg.yml")

	flag.Parse()
	Init()

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	originalListener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ServerHost, ServerPort))
	if err != nil {
		panic(err)
	}

	sl, err := stoppableListener.New(originalListener)
	if err != nil {
		panic(err)
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/jpeg", handlers.JpegHandler)
	http.HandleFunc("/jpeg/", handlers.JpegHandler)
	http.HandleFunc("/mjpeg", handlers.MjpegHandler)
	http.HandleFunc("/mjpeg/", handlers.MjpegHandler)
	http.HandleFunc("/mjpeg/viewer", handlers.MjpegViewerHandler)
	http.HandleFunc("/mjpeg/viewer/", handlers.MjpegViewerHandler)
	http.HandleFunc("/stop", handlers.StopHandler)
	http.HandleFunc("/stop/", handlers.StopHandler)

	server := http.Server{}

	var swg sync.WaitGroup
	go func() {
		swg.Add(1)
		defer swg.Done()
		server.Serve(sl)
	}()

	screenshot.InitConn()

	if Mode == "single" {
		go func() {
			std_interval := float64(1.0 / float64(Fps))
			time_to_sleep := std_interval
			s := time.Now()
			for {
				if Done {
					return
				}
				img := CaptureWindowMust(&screenshot.POS{Left, Top},
					&screenshot.SIZE{Width, Height},
					&screenshot.RESIZE{ResizeWidth, ResizeHeight},
					ToSBS,
					Cursor)
				sio := stringio.New()
				err = ljpeg.Encode(sio, img, &ljpeg.EncoderOptions{Quality: Quality})
				if err == nil {
					sio.Seek(0, 0)
					handlers.Images <- sio
				}
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
			}
		}()
	} else if Mode == "single-single" {
		go func() {
			std_interval := float64(1.0 / float64(Fps))
			time_to_sleep := std_interval
			s := time.Now()
			for {
				if Done {
					return
				}
				img := CaptureWindowMust(&screenshot.POS{Left, Top},
					&screenshot.SIZE{Width, Height},
					&screenshot.RESIZE{ResizeWidth, ResizeHeight},
					ToSBS,
					Cursor)
				handlers.Buffer <- img
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
			}
		}()

		go func() {
			for {
				if Done {
					return
				}
				img := <-handlers.Buffer
				sio := stringio.New()
				err = ljpeg.Encode(sio, img, &ljpeg.EncoderOptions{Quality: Quality})
				if err == nil {
					sio.Seek(0, 0)
					handlers.Images <- sio
				}
			}
		}()
	} else if Mode == "single-multi" {
		go func() {
			std_interval := float64(1.0 / float64(Fps))
			time_to_sleep := std_interval
			s := time.Now()
			for {
				if Done {
					return
				}
				img := CaptureWindowMust(&screenshot.POS{Left, Top},
					&screenshot.SIZE{Width, Height},
					&screenshot.RESIZE{ResizeWidth, ResizeHeight},
					ToSBS,
					Cursor)
				handlers.Buffer <- img
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
			}
		}()

		for i := 0; i < Convert; i += 1 {
			go func(id int) {
				Log.Info(fmt.Sprintf("converter: %d", id))
				for {
					if Done {
						return
					}
					img := <-handlers.Buffer
					sio := stringio.New()
					err = ljpeg.Encode(sio, img, &ljpeg.EncoderOptions{Quality: Quality})
					if err == nil {
						sio.Seek(0, 0)
						handlers.Images <- sio
					}
				}
			}(i)
		}
	} else if Mode == "multi-multi" {
		for i := 0; i < Convert; i += 1 {
			go func(id int) {
				Log.Info(fmt.Sprintf("converter: %d", id))
				std_interval := float64(1.0 / float64(Fps))
				time_to_sleep := std_interval
				s := time.Now()
				for {
					if Done {
						return
					}
					img := CaptureWindowMust(&screenshot.POS{Left, Top},
						&screenshot.SIZE{Width, Height},
						&screenshot.RESIZE{ResizeWidth, ResizeHeight},
						ToSBS,
						Cursor)
					sio := stringio.New()
					err = ljpeg.Encode(sio, img, &ljpeg.EncoderOptions{Quality: Quality})
					if err == nil {
						sio.Seek(0, 0)
						handlers.Images <- sio
					}
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
				}
			}(i)
		}
	} else if Mode == "sync-single-multi" {
		go func() {
			n := 0
			first := true
			std_interval := float64(1.0 / float64(Fps))
			time_to_sleep := std_interval
			s := time.Now()
			for {
				if Done {
					return
				}
				if n == Convert {
					if first {
						first = false
					}
					n = 0
				}
				img := CaptureWindowMust(&screenshot.POS{Left, Top},
					&screenshot.SIZE{Width, Height},
					&screenshot.RESIZE{ResizeWidth, ResizeHeight},
					ToSBS,
					Cursor)
				handlers.BufferArray[n] <- img
				if !first {
					sio := <-handlers.ImagesArray[n]
					handlers.Images <- sio
				}
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
				n += 1
			}
		}()

		for i := 0; i < Convert; i += 1 {
			go func(id int) {
				Log.Info(fmt.Sprintf("converter: %d", id))
				for {
					if Done {
						return
					}
					img := <-handlers.BufferArray[id]
					sio := stringio.New()
					err = ljpeg.Encode(sio, img, &ljpeg.EncoderOptions{Quality: Quality})
					if err == nil {
						sio.Seek(0, 0)
						handlers.ImagesArray[id] <- sio
					}
				}
			}(i)
		}
	} else if Mode == "sync-multi-multi" {
		go func() {
			n_shot := 0
			n_convert := 0
			first_shot := true
			first_convert := true
			std_interval := float64(1.0 / float64(Fps))
			time_to_sleep := std_interval
			s := time.Now()
			for {
				if Done {
					return
				}
				if n_shot == Shot {
					if first_shot {
						first_shot = false
					}
					n_shot = 0
				}
				if n_convert == Convert {
					if first_convert {
						first_convert = false
					}
					n_convert = 0
				}
				handlers.TasksArray[n_shot] <- n_convert

				if !first_convert {
					sio := <-handlers.ImagesArray[n_convert]
					handlers.Images <- sio
				}
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
				n_shot += 1
				n_convert += 1
			}
		}()

		for i := 0; i < Shot; i += 1 {
			go func(id int) {
				Log.Info(fmt.Sprintf("shotter: %d", id))
				for {
					if Done {
						return
					}
					convert_id := <-handlers.TasksArray[id]
					img := CaptureWindowMust(&screenshot.POS{Left, Top},
						&screenshot.SIZE{Width, Height},
						&screenshot.RESIZE{ResizeWidth, ResizeHeight},
						ToSBS,
						Cursor)
					handlers.BufferArray[convert_id] <- img
				}
			}(i)
		}

		for i := 0; i < Convert; i += 1 {
			go func(id int) {
				Log.Info(fmt.Sprintf("converter: %d", id))
				for {
					if Done {
						return
					}
					img := <-handlers.BufferArray[id]
					sio := stringio.New()
					err = ljpeg.Encode(sio, img, &ljpeg.EncoderOptions{Quality: Quality})
					// err = jpeg.Encode(sio, img, &jpeg.Options{Quality})
					if err == nil {
						sio.Seek(0, 0)
						handlers.ImagesArray[id] <- sio
					}
				}
			}(i)
		}
	}

	Log.Info(fmt.Sprintf("Serving HTTP"))

	select {
	case signal := <-sigs:
		Log.Info(fmt.Sprintf("Got signal: (%v)", signal))
	case _ = <-handlers.Done:
		Log.Info(fmt.Sprintf("Stop by /stop handler"))
	}

	defer Log.Flush()
	screenshot.CloseConn()
	Done = true
	Log.Info(fmt.Sprintf("Stopping server"))
	Log.Info(fmt.Sprintf("Server will stop in 5 seconds ..."))
	time.Sleep(5)
	sl.Stop()
	Log.Info(fmt.Sprintf("Waiting on server"))
	swg.Wait()
	Log.Info("Server exit!")
	os.Exit(0)
}
