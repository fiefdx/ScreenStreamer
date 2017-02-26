package rtmp

import (
	"time"

	"flv.go/flv"
)

var Buffer chan *flv.AVCVideoFrame
var FrameBuffer *flv.AVCVideoFrame
var Fps int
var Alpha int

func GetFrame() *flv.AVCVideoFrame {
	select {
	case frame := <-Buffer:
		FrameBuffer = frame
		return FrameBuffer
	default:
		return FrameBuffer
	}
}

func GetFrameMust() *flv.AVCVideoFrame {
	frame := GetFrame()
	for frame == nil {
		frame = GetFrame()
	}
	return frame
}

func InitBuf(b_size int) {
	Buffer = make(chan *flv.AVCVideoFrame, b_size)
}

func InitCap(fps, alpha int) {
	Fps = fps
	Alpha = alpha
}

type ServerHandler interface {
	OnPublishing(s *RtmpNetStream) error
	OnPlaying(s *RtmpNetStream) error
	OnClosed(s *RtmpNetStream)
	OnError(s *RtmpNetStream, err error)
}

type ClientHandler interface {
	OnPublishStart(s *RtmpNetStream) error
	OnPlayStart(s *RtmpNetStream) error
	OnClosed(s *RtmpNetStream)
	OnError(s *RtmpNetStream, err error)
}

type DefaultClientHandler struct {
}

func (this *DefaultClientHandler) OnPublishStart(s *RtmpNetStream) error {
	return nil
}
func (this *DefaultClientHandler) OnPlayStart(s *RtmpNetStream) error {
	if obj, found := findObject(s.streamName); !found {
		obj, err := new_streamObject(s.streamName, 90*time.Second, true, 0) // 10
		if err != nil {
			return err
		}
		s.obj = obj
	} else {
		s.obj = obj
	}
	return nil
}
func (this *DefaultClientHandler) OnClosed(s *RtmpNetStream) {
	if s.mode == MODE_PRODUCER {
		log.Infof("RtmpNetStream Publish %s %s closed", s.conn.remoteAddr, s.path)
		if obj, found := findObject(s.streamName); found {
			obj.Close()
		}
	}
}
func (this *DefaultClientHandler) OnError(s *RtmpNetStream, err error) {
	log.Errorf("RtmpNetStream %v %s %s %+v", s.mode, s.conn.remoteAddr, s.path, err)
	s.Close()
}

func ScreenShotService(s *RtmpNetStream) {
	n := 0
	ttt := time.Now()
	// std_interval := float64(1.0 / float64(Fps))
	// time_to_sleep := std_interval
	for {
		if n == 30 {
			tt := time.Now()
			fps := float64(30) / tt.Sub(ttt).Seconds()
			log.Warnf(">>>>>>  FPS: %.2ffps <<<<<<", fps)
			n = 0
			ttt = time.Now()
		}

		frame := <-Buffer

		if frame.GetType() == flv.TAG_TYPE_VIDEO {
			payload := make([]byte, 0)
			payload = append(payload, *frame.GetBody()...)

			sp := NewMediaFrame()
			sp.Timestamp = frame.GetDts()
			sp.Type = byte(frame.GetType())
			sp.StreamId = frame.GetStream()

			one := payload[0]
			sp.VideoFrameType = one >> 4
			sp.VideoCodecID = one & 0x0f

			sp.Payload.Write(payload)
			f, e := s.obj.WriteFrame(sp)
			for f != true {
				time.Sleep(time.Millisecond * 5)
				f, e = s.obj.WriteFrame(sp)
			}
			if e != nil {
				log.Warn(e)
			}
		}

		n += 1
	}
}

type DefaultServerHandler struct {
}

func (p *DefaultServerHandler) OnPublishing(s *RtmpNetStream) error {
	if obj, found := findObject(s.streamName); !found {
		obj, err := new_streamObject(s.streamName, 90*time.Second, true, 1) // 10
		if err != nil {
			return err
		}
		s.obj = obj
		if s.streamName == "live/screen" {
			go ScreenShotService(s)
		}
	} else {
		s.obj = obj
	}
	return nil
}

func (p *DefaultServerHandler) OnPlaying(s *RtmpNetStream) error {
	if obj, found := findObject(s.streamName); !found {
		obj, err := new_streamObject(s.streamName, 90*time.Second, true, 0) // 10
		if err != nil {
			return err
		}
		s.obj = obj
	} else {
		s.obj = obj
	}
	s.obj.Attach(s)
	go s.writeLoop()
	return nil
}

func (p *DefaultServerHandler) OnClosed(s *RtmpNetStream) {
	// mode := "UNKNOWN"
	// if s.mode == MODE_CONSUMER {
	// 	mode = "CONSUMER"
	// } else if s.mode == MODE_PROXY {
	// 	mode = "PROXY"
	// } else if s.mode == MODE_CONSUMER|MODE_PRODUCER {
	// 	mode = "PRODUCER|CONSUMER"
	// }
	// log.Infof("RtmpNetStream %v %s %s closed", mode, s.conn.remoteAddr, s.path)
	// if d, ok := find_broadcast(s.path); ok {
	// 	if s.mode == MODE_PRODUCER {
	// 		d.stop()
	// 	} else if s.mode == MODE_CONSUMER {
	// 		d.removeConsumer(s)
	// 	} else if s.mode == MODE_CONSUMER|MODE_PRODUCER {
	// 		d.removeConsumer(s)
	// 		d.stop()
	// 	}
	// }
	if s.mode == MODE_PRODUCER {
		log.Infof("RtmpNetStream Publish %s %s closed", s.conn.remoteAddr, s.path)
		if obj, found := findObject(s.streamName); found {
			obj.Close()
		}
	}
}

func (p *DefaultServerHandler) OnError(s *RtmpNetStream, err error) {
	log.Errorf("RtmpNetStream %s %s %+v", s.conn.remoteAddr, s.path, err)
	s.Close()
}
