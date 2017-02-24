package codec

/*
#cgo LDFLAGS: -lavcodec -lavformat -lavutil
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
#include <libavutil/opt.h>

AVCodec *codec;

typedef struct {
	int w, h;
	int pixfmt;
	char *preset[2];
	char *profile;
	int bitrate;
	int got;
	AVCodec *c;
	AVCodecContext *ctx;
	AVFrame *f;
	AVPacket pkt;
	int gop_size;
	int sample_rate;
	int time_base_den;
	int time_base_num;
} h264enc_t;

static int h264enc_new(h264enc_t *m) {
	m->c = avcodec_find_encoder(AV_CODEC_ID_H264); // AV_CODEC_ID_MPEG1VIDEO CODEC_ID_H264 CODEC_ID_MPEG4
	m->ctx = avcodec_alloc_context3(m->c);
	m->ctx->width = m->w;
	m->ctx->height = m->h;
	m->ctx->bit_rate = m->bitrate;
	// m->ctx->bit_rate = 1024 * 1024 * 1024;
	m->ctx->pix_fmt = m->pixfmt;
	m->ctx->flags |= CODEC_FLAG_GLOBAL_HEADER;
	m->ctx->time_base.num = m->time_base_num; // 1;
	m->ctx->time_base.den = m->time_base_den; // 10;
	m->f = av_frame_alloc();
	av_opt_set(m->ctx->priv_data, "preset", "ultrafast", 0);
	av_opt_set(m->ctx->priv_data, "tune", "zerolatency", 0);
	m->c->capabilities &= ~CODEC_CAP_DELAY;
	m->ctx->delay = 0;
	// m->ctx->keyint_min = 30;
	m->ctx->gop_size = m->gop_size; // 10;
	m->ctx->sample_rate = m->sample_rate; // 10;
	m->f->width = m->w;
	m->f->height = m->h;
	m->got = 0;
	return avcodec_open2(m->ctx, m->c, NULL);
}

static void init_codec(AVCodec *c) {
	c->capabilities &= ~CODEC_CAP_DELAY;
}

static void init_ctx(AVCodecContext *ctx, int w, int h, int pixfmt, int time_base_num, int time_base_den, int gop_size, int sample_rate) {
	ctx->width = w;
	ctx->height = h;
	ctx->bit_rate = 0;
	ctx->pix_fmt = pixfmt;
	ctx->flags |= CODEC_FLAG_GLOBAL_HEADER;
	ctx->time_base.num = time_base_num; // 1;
	ctx->time_base.den = time_base_den; // 10;
	av_opt_set(ctx->priv_data, "preset", "ultrafast", 0);
	av_opt_set(ctx->priv_data, "tune", "zerolatency", 0);
	ctx->delay = 0;
	// ctx->keyint_min = 30;
	ctx->gop_size = gop_size; // 10;
	ctx->sample_rate = sample_rate; // 10;
}

static void init_frame(AVFrame *f, int w, int h) {
	f->width = w;
	f->height = h;
}
*/
import "C"

import (
	"errors"
	// "fmt"
	"image"
	"strings"
	"unsafe"
	//"log"
)

type H264Encoder struct {
	m          C.h264enc_t
	Header     []byte
	Pixfmt     image.YCbCrSubsampleRatio
	W, H       int
	FramesPass int
}

type H264Encoder2 struct {
	W, H   int
	Header []byte
	Pixfmt image.YCbCrSubsampleRatio
	Codec  *C.AVCodec
	Ctx    *C.AVCodecContext
	Frame  *C.AVFrame
	Pkt    *C.AVPacket
	Got    C.int
}

func NewH264Encoder(
	w, h,
	gop_size,
	sample_rate,
	time_base_num,
	time_base_den,
	bit_rate int,
	pixfmt image.YCbCrSubsampleRatio,
	opts ...string,
) (m *H264Encoder, err error) {
	m = new(H264Encoder)
	m.FramesPass = 0
	m.m.w = (C.int)(w)
	m.m.h = (C.int)(h)
	m.W = w
	m.H = h
	m.m.bitrate = 1024 * 1024 * (C.int)(bit_rate) // 1024 * 1024 * 64 // 1024 * 1024 * 128 // 1024 * 1024 * 256 // 1024 * 1024 * 512
	m.m.gop_size = (C.int)(gop_size)
	m.m.sample_rate = (C.int)(sample_rate)
	m.m.time_base_num = (C.int)(time_base_num)
	m.m.time_base_den = (C.int)(time_base_den)
	m.Pixfmt = pixfmt
	switch pixfmt {
	case image.YCbCrSubsampleRatio444:
		m.m.pixfmt = C.PIX_FMT_YUV444P
	case image.YCbCrSubsampleRatio422:
		m.m.pixfmt = C.PIX_FMT_YUV422P
	case image.YCbCrSubsampleRatio420:
		m.m.pixfmt = C.PIX_FMT_YUV420P
	}
	for _, opt := range opts {
		a := strings.Split(opt, ",")
		switch {
		case a[0] == "preset" && len(a) == 3:
			m.m.preset[0] = C.CString(a[1])
			m.m.preset[1] = C.CString(a[2])
		case a[0] == "profile" && len(a) == 2:
			m.m.profile = C.CString(a[1])
		}
	}
	// fmt.Printf("\n>>>>>> m.m.pixfmt: %v\n", m.m.pixfmt)
	r := C.h264enc_new(&m.m)
	if int(r) < 0 {
		err = errors.New("open encoder failed")
		return
	}
	m.Header = fromCPtr(unsafe.Pointer(m.m.ctx.extradata), (int)(m.m.ctx.extradata_size))
	// m.Header = fromCPtr(unsafe.Pointer(m.m.pps), (int)(m.m.ppslen))
	return
}

func NewH264Encoder2(
	w, h,
	gop_size,
	sample_rate,
	time_base_num,
	time_base_den int,
	pixfmt image.YCbCrSubsampleRatio,
	opts ...string,
) (m *H264Encoder2, err error) {
	m = &H264Encoder2{}
	m.W = w
	m.H = h
	m.Codec = C.avcodec_find_encoder(C.AV_CODEC_ID_H264)
	C.init_codec(m.Codec)
	cpixfmt := 0
	switch pixfmt {
	case image.YCbCrSubsampleRatio444:
		cpixfmt = C.PIX_FMT_YUV444P
	case image.YCbCrSubsampleRatio422:
		cpixfmt = C.PIX_FMT_YUV422P
	case image.YCbCrSubsampleRatio420:
		cpixfmt = C.PIX_FMT_YUV420P
	}
	m.Ctx = C.avcodec_alloc_context3(m.Codec)
	C.init_ctx(m.Ctx, (C.int)(w), (C.int)(h), (C.int)(cpixfmt), (C.int)(time_base_num), (C.int)(time_base_den), (C.int)(gop_size), (C.int)(sample_rate))
	m.Frame = C.av_frame_alloc()
	m.Pixfmt = pixfmt
	C.init_frame(m.Frame, (C.int)(w), (C.int)(h))

	r := C.avcodec_open2(m.Ctx, m.Codec, nil)
	if int(r) < 0 {
		err = errors.New("open encoder failed")
		return
	}
	m.Header = fromCPtr(unsafe.Pointer(m.Ctx.extradata), (int)(m.Ctx.extradata_size))
	return
}

type H264Out struct {
	Data []byte
	Key  bool
}

func (m *H264Encoder) Encode(img *image.YCbCr) (out H264Out, err error) {
	var f *C.AVFrame
	if img == nil {
		f = nil
	} else {
		if img.SubsampleRatio != m.Pixfmt {
			err = errors.New("image pixfmt not match")
			return
		}
		if img.Rect.Dx() != m.W || img.Rect.Dy() != m.H {
			err = errors.New("image size not match")
			return
		}
		if m.FramesPass > 30 {
			m.m.f = C.av_frame_alloc()
			m.FramesPass = 0
		}
		m.FramesPass += 1
		f = m.m.f
		f.data[0] = (*C.uint8_t)(unsafe.Pointer(&img.Y[0]))
		f.data[1] = (*C.uint8_t)(unsafe.Pointer(&img.Cb[0]))
		f.data[2] = (*C.uint8_t)(unsafe.Pointer(&img.Cr[0]))

		f.linesize[0] = (C.int)(img.YStride)
		f.linesize[1] = (C.int)(img.CStride)
		f.linesize[2] = (C.int)(img.CStride)
		f.width = (C.int)(m.W)
		f.height = (C.int)(m.H)
		f.format = 5 // 1 // AV_PIX_FMT_YUV420P // 5 // AV_PIX_FMT_YUV444P
	}

	C.av_init_packet(&m.m.pkt)
	// m.m.pkt.data = C.NULL
	m.m.pkt.size = 0
	r := C.avcodec_encode_video2(m.m.ctx, &m.m.pkt, f, &m.m.got)
	defer C.av_free_packet(&m.m.pkt)
	if int(r) < 0 {
		err = errors.New("encode failed")
		return
	}
	if m.m.got == 0 {
		err = errors.New("no picture")
		return
	}
	if m.m.pkt.size == 0 {
		err = errors.New("packet size == 0")
		return
	}

	out.Data = make([]byte, m.m.pkt.size)
	C.memcpy(
		unsafe.Pointer(&out.Data[0]),
		unsafe.Pointer(m.m.pkt.data),
		(C.size_t)(m.m.pkt.size),
	)
	out.Key = (m.m.pkt.flags & C.AV_PKT_FLAG_KEY) != 0

	return
}

func (m *H264Encoder2) Encode(img *image.YCbCr) (out H264Out, err error) {
	var f *C.AVFrame
	if img == nil {
		f = nil
	} else {
		if img.SubsampleRatio != m.Pixfmt {
			err = errors.New("image pixfmt not match")
			return
		}
		if img.Rect.Dx() != m.W || img.Rect.Dy() != m.H {
			err = errors.New("image size not match")
			return
		}
		f = m.Frame
		f.data[0] = (*C.uint8_t)(unsafe.Pointer(&img.Y[0]))
		f.data[1] = (*C.uint8_t)(unsafe.Pointer(&img.Cb[0]))
		f.data[2] = (*C.uint8_t)(unsafe.Pointer(&img.Cr[0]))
		f.linesize[0] = (C.int)(img.YStride)
		f.linesize[1] = (C.int)(img.CStride)
		f.linesize[2] = (C.int)(img.CStride)
		// f.width = 1920
		// f.height = 1080
		f.format = 5 // 1 // AV_PIX_FMT_YUV420P // 5 // AV_PIX_FMT_YUV444P
	}

	C.av_init_packet(m.Pkt)
	// m.m.pkt.data = C.NULL
	m.Pkt.size = 0
	r := C.avcodec_encode_video2(m.Ctx, m.Pkt, f, (*C.int)(&m.Got))
	defer C.av_free_packet(m.Pkt)
	if int(r) < 0 {
		err = errors.New("encode failed")
		return
	}
	if m.Got == 0 {
		err = errors.New("no picture")
		return
	}
	if m.Pkt.size == 0 {
		err = errors.New("packet size == 0")
		return
	}

	out.Data = make([]byte, m.Pkt.size)
	C.memcpy(
		unsafe.Pointer(&out.Data[0]),
		unsafe.Pointer(m.Pkt.data),
		(C.size_t)(m.Pkt.size),
	)
	out.Key = (m.Pkt.flags & C.AV_PKT_FLAG_KEY) != 0

	return
}
