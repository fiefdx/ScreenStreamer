ScreenStreamer
==============
A golang project, aim for stream one pc's active window or full screen (windows or linux) to another device as MJPEG
over http or as flv over rtmp as soon as possible, it can be a broadcast service or a point to point service as you
want, if you want to use tridef-3d but you don't have a professional VR head set, and you have a smart phone, I think
you will like this application.

For using ScreenStreamer easily, you can use [ScreenStreamerManager](https://github.com/fiefdx/ScreenStreamerManager).

With rtmp, you should wait a few minutes, to get the video delay down to 500ms. With Windows this not work, With Windows,
you can minimize the window temporarily and wait to the phone's screen stopped, then restore the window.

ScreenShots
-----------
1. Working like this
   
   ![Alt text](/doc/stream_video_to_phone.png?raw=true "stream_video_to_phone")

Install & Run ScreenStreamer
----------------------------
1. Build it
   
   ```bash
   # need install libjpeg/libjpeg-turbo, on windows, need install mingw, libjpeg/libjpeg-turbo for mjpeg
   sudo apt-get install libjpeg-turbo8-dev
   # need install libav, on windows, need install mingw, libav for rtmp
   sudo apt-get install libav-tools libavcodec-dev libavformat-dev
   # create a workspace directory
   mkdir ./ScreenStreamer
   # enter the directory
   cd ./ScreenStreamer
   # clone the source code
   git clone git@github.com:fiefdx/ScreenStreamer.git src/ScreenStreamer
   # enter the project root directory
   cd ./src/ScreenStreamer
   # add the workspace directory and lib directory to GOPATH
   # export GOPATH=/path-to-workspace/ScreenStreamer:/path-to-workspace/ScreenStreamer/src/ScreenStreamer/lib
   # run . ./dev.sh can set the GOPATH too
   . ./dev.sh
   # build the mjpeg, it will produce binary executable files named "mjpeg" or "mjpeg.exe"
   go build ./cmd/mjpeg/mjpeg.go
   # build the rtmp, it will produce binary executable files named "rtmp" or "rtmp.exe"
   go build ./cmd/rtmp/rtmp.go
   # build the get_active_window, it can get current active window's id for ScreenStreamerManager
   go build ./cmd/get_active_window/get_active_window.go

   ```
2. Config configuration file
   
   ```yaml
   configuration.mjpeg.yml for mjpeg
   # log level
   # a value of (DEBUG, INFO, WARN, ERROR)
   log_level: INFO
   log_path: logs

   # host and port that you will use other device to request this service
   server_host: 0.0.0.0
   server_port: 8080

   # how many threads you want to run goroutines
   threads: 4

   # target fps
   fps: 23
   # target jpeg quality
   quality: 97
   # broadcast,true for broadcast, false for point to point
   broadcast: false
   # screenshot image 
   # left offset, 0 is default
   # top offset, 0 is default
   # width, 0 is default, means auto
   # height, 0 is default, means auto
   left: 0
   top: 0
   width: 0
   height: 0
   # resize function is not work now, so ignore this two options
   resize_width: 0
   resize_height: 0

   # don't change this options, if you don't know how it works
   tasks_queue_size: 1
   buffer_queue_size: 0
   images_queue_size: 0
   convert_buffer_size: 1
   convert_images_size: 1

   # to_sbs, convert 2d image to side-by-side images
   to_sbs: false
   # cursor, capture cursor or not
   cursor: false

   # how many screen shoters to work, less than 10
   shot: 1
   # alpha is an option to make the fps stable
   # every frame sleep +/- (alpha / 100) * (1 / fps) senonds
   alpha: 15
   # how many image converter to work, less than 10
   convert: 3
   # the application work mode
   # single, just a worker with shoter & converter functions
   # single-single, a shoter and a converter
   # single-multi, a shoter and multi converters
   # multi-multi, multi shoter and multi converters
   # sync-single-multi, a shoter and multi converters, shoter and converters works synchronizely
   # sync-multi-multi, multi shoters and multi converters, shoters and converters works synchronizely
   mode: single


   configuration.rtmp.yml for rtmp
   # it not support broadcast, currently
   # log level
   # a value of (DEBUG, INFO, WARN, ERROR)
   log_level: INFO
   log_path: logs

   # host and port that you will use other device to request this service
   server_host: 0.0.0.0
   server_port: 1935

   # how many threads you want to run goroutines
   threads: 5

   # target fps
   fps: 30

   # write_frames_timeout, write frames deadline duration, default is 500ms
   write_frames_timeout: 500

   # bit_rate, default is 64M
   bit_rate: 64

   # full_screen, default is true
   full_screen: true # windowed mode not work now, so just set to true, if you want to capture window, then use ScreenStreamerManager

   # screenshot image
   # left offset, 0 is default
   # top offset, 0 is default
   # width, 0 is default, means auto
   # height, 0 is default, means auto
   left: 0
   top: 0
   width: 0
   height: 0
   # resize function is not work now, so ignore this two options
   resize_width: 0
   resize_height: 0

   # don't change this options, if you don't know how it works
   buffer_queue_size: 0

   # to_sbs, convert 2d image to side-by-side images, it's not work on full screen mode
   to_sbs: false
   # cursor, capture cursor or not
   cursor: false

   # alpha is an option to make the fps stable
   # every frame sleep +/- (alpha / 100) * (1 / fps) senonds
   alpha: 15
   # how many YCbCr converter to work for one image
   convert: 4
   # the application work mode
   # single, just a worker with shoter & converter functions
   mode: single
   ```
3. Run it
   
   ```bash
   # enter the project root directory
   cd ./src/ScreenStreamer
   # run mjpeg service
   ./mjpeg or .\mjpeg.exe
   # use a web browser or other video player, open http://host:port/mjpeg

   # run rtmp service
   ./rtmp or .\rtmp.exe
   # use a video player, open rtmp://host:port/live/screen
   ```
4. Stop it
   
   ```bash
   # Use Ctrl+C at the terminal or just close the terminal
   ```
5. Performance
   
   ```
   Server side
   CPU: i5-6300HQ(20%)
   RAM: 12GB(very less)
   SYSTEM: Windows 10 Home

   As MJPEG: # mxplayer can't play mjpeg sometimes
   Client side
   PHONE: LG G3
   PLAYER: mxplayer
   FPS: 25(the mxplayer play mjpeg as 25fps,so , I set it 20-24fps, more fps just get frame delay, and the window's size is (1280 * 720))
   DELAY: <100ms(it's depend on your network and devices, my phone can's fully use 1000M wifi)

   As RTMP:
   Client side
   PHONE: LG G3
   PLAYER: mxplayer
   FPS: 20(my laptop is 1920*1080 resolution, it can get 20fps in Windows 10, or 25fps in Xubuntu linux)
   DELAY: <500ms(it's depend on your network and devices, my phone and pc not powerful enough)

   Wifi: 1000M
   ```
