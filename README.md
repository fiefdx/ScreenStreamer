ScreenStreamer
==============
A golang project, aim for stream one pc's active window (windows or linux) to another device as MJPEG over http as soon as possible,
it's not a broadcast server, it's for point to point service, if you want to use tridef-3d and you don't have a professional VR head
set, and you have a smart phone, I think you will like this application.

Install & Run ScreenStreamer
----------------------------
1. Build it
   
   ```bash
   # create a workspace directory
   mkdir ./ScreenStreamer
   # enter the directory
   cd ./ScreenStreamer
   # clone the source code
   git clone git@github.com:fiefdx/ScreenStreamer.git src/ScreenStreamer
   # enter the project root directory
   cd ./src/ScreenStreamer
   # add the workspace directory and lib directory to GOPATH
   export GOPATH=/path-to-workspace/ScreenStreamer:/path-to-workspace/ScreenStreamer/src/ScreenStreamer/lib
   # build the project, it will produce a binary executable file named "stream" or "stream.exe"
   go build stream.go

   ```
2. Config configuration.yml file
   
   ```yaml
   # log level
   # a value of (DEBUG, INFO, WARN, ERROR)
   log_level: INFO
   log_path: logs

   # host and port that you will use other device to request this service
   server_host: 0.0.0.0
   server_port: 8080

   # how many threads you want to run goroutines
   threads: 5

   # target fps
   fps: 23
   # target jpeg quality
   quality: 96
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
   mode: sync-single-multi

   # I use it to stream windows window to my smartphone, play the stream with mxplayer, the mxplayer play mjpeg as 25fps,
   # so, I set it 22-24fps, and the window's size is (1280 * 720), I use 1000M wifi.
   ```
3. Run it
   
   ```bash
   # enter the project root directory
   cd ./src/ScreenStreamer
   # run it
   ./stream or .\stream.exe
   # use a web browser or other video player, open http://host:port/mjpeg
   ```
4. Stop it
   
   ```bash
   # Use Ctrl+C at the terminal or 
   # Use a web browser or curl open http://host:port/stop
   curl host:port/stop
   ```
5. Performance
   
   ```
   Server side
   CPU: i5-6300HQ(40%)
   RAM: 12GB(very less)
   SYSTEM: Windows 10 Home

   Client side
   PHONE: LG G3
   PLAYER: mxplayer
   FPS: 25
   DELAY: 100-500ms(it's depend on your network and devices, my phone can's fully use 1000M wifi)

   Wifi: 1000M
   ```
