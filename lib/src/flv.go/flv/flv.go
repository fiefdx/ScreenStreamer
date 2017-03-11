package flv

import (
	"bytes"
	"fmt"
	"github.com/metachord/amf.go/amf0"
	"io"
	"os"
	"stringio"
	"time"
)

type Header struct {
	Version uint16
	Body    []byte
}

type Frame interface {
	WriteFrame(io.Writer) error
	GetBody() *[]byte
	GetStream() uint32
	GetDts() uint32
	SetDts(dts uint32)
	GetType() TagType
	GetPrevTagSize() uint32
	String() string
}

type CFrame struct {
	Stream      uint32
	Dts         uint32
	Type        TagType
	Flavor      Flavor
	Position    int64
	Body        []byte
	PrevTagSize uint32
}

type VideoFrame struct {
	*CFrame
	CodecId VideoCodec
	Width   uint16
	Height  uint16
}

type AVCVideoFrame struct {
	*VideoFrame
	PacketType AvcPacketType
	StartTime  time.Time
}

type AudioFrame struct {
	*CFrame
	CodecId  AudioCodec
	Rate     uint32
	BitSize  AudioSize
	Channels AudioType
}

type MetaFrame struct {
	*CFrame
}

func (f *CFrame) WriteFrame(w io.Writer) error {
	bl := uint32(len(f.Body))
	var err error
	err = writeType(w, f.Type)
	if err != nil {
		return err
	}
	err = writeBodyLength(w, bl)
	err = writeDts(w, f.Dts)
	err = writeStream(w, f.Stream)
	err = writeBody(w, f.Body)
	prevTagSize := bl + uint32(TAG_HEADER_LENGTH)
	err = writePrevTagSize(w, prevTagSize)
	return nil
}

func (f *CFrame) GetBody() *[]byte {
	return &f.Body
}
func (f *CFrame) GetStream() uint32 {
	return f.Stream
}
func (f *CFrame) GetDts() uint32 {
	return f.Dts
}
func (f *CFrame) SetDts(dts uint32) {
	f.Dts = dts
}
func (f *CFrame) GetType() TagType {
	return f.Type
}
func (f *CFrame) GetPrevTagSize() uint32 {
	return f.PrevTagSize
}

func writeType(w io.Writer, t TagType) error {
	_, err := w.Write([]byte{byte(t)})
	return err
}

func writeBodyLength(w io.Writer, bl uint32) error {
	_, err := w.Write([]byte{byte(bl >> 16), byte((bl >> 8) & 0xFF), byte(bl & 0xFF)})
	return err
}

func writeDts(w io.Writer, dts uint32) error {
	_, err := w.Write([]byte{byte((dts >> 16) & 0xFF), byte((dts >> 8) & 0xFF), byte(dts & 0xFF)})
	if err != nil {
		return err
	}
	_, err = w.Write([]byte{byte((dts >> 24) & 0xFF)})
	return err
}

func writeStream(w io.Writer, stream uint32) error {
	_, err := w.Write([]byte{byte(stream >> 16), byte((stream >> 8) & 0xFF), byte(stream & 0xFF)})
	return err
}

func writeBody(w io.Writer, body []byte) error {
	_, err := w.Write(body)
	return err
}

func writePrevTagSize(w io.Writer, prevTagSize uint32) error {
	_, err := w.Write([]byte{byte((prevTagSize >> 24) & 0xFF), byte((prevTagSize >> 16) & 0xFF), byte((prevTagSize >> 8) & 0xFF), byte(prevTagSize & 0xFF)})
	return err
}

func (f VideoFrame) String() string {
	s := ""
	switch f.Flavor {
	case KEYFRAME:
		s = fmt.Sprintf("%10d\t%d\t%d\t%s\t%s\t{%dx%d,%d bytes} keyframe", f.CFrame.Stream, f.CFrame.Dts, f.CFrame.Position, f.CFrame.Type, f.CodecId, f.Width, f.Height, len(f.CFrame.Body))
	default:
		s = fmt.Sprintf("%10d\t%d\t%d\t%s\t%s\t{%dx%d,%d bytes}", f.CFrame.Stream, f.CFrame.Dts, f.CFrame.Position, f.CFrame.Type, f.CodecId, f.Width, f.Height, len(f.CFrame.Body))
	}
	return s
}

func (f AVCVideoFrame) String() string {
	s := ""
	s = fmt.Sprintf("%10d\t%d\t%d\t%s\t%s\t{%s,%dx%d,%d bytes}", f.CFrame.Stream, f.CFrame.Dts, f.CFrame.Position, f.CFrame.Type, f.CodecId, f.PacketType, f.Width, f.Height, len(f.CFrame.Body))
	if f.Flavor == KEYFRAME {
		s += " seekable"
	}
	return s
}

func (f AudioFrame) String() string {
	return fmt.Sprintf("%10d\t%d\t%d\t%s\t%s\t{%d,%s,%s,%d bytes}", f.CFrame.Stream, f.CFrame.Dts, f.CFrame.Position, f.CFrame.Type, f.CodecId, f.Rate, f.BitSize, f.Channels, len(f.CFrame.Body))
}

func (f MetaFrame) String() string {
	buf := bytes.NewReader(f.CFrame.Body)
	dec := amf0.NewDecoder(buf)
	evName, err := dec.Decode()
	mds := ""
	if err == nil {
		switch evName {
		case amf0.StringType("onMetaData"):
			md, err := dec.Decode()
			if err == nil {

				var ea map[amf0.StringType]interface{}
				switch md := md.(type) {
				case *amf0.EcmaArrayType:
					ea = *md
				case *amf0.ObjectType:
					ea = *md
				}
				for k, v := range ea {
					mds += fmt.Sprintf("%v=%+v;", k, v)
				}
			}
		}
	}

	return fmt.Sprintf("%10d\t%d\t%d\t%s\t%s", f.CFrame.Stream, f.CFrame.Dts, f.CFrame.Position, f.CFrame.Type, mds)
}

type FlvReader struct {
	InFile *os.File
	width  uint16
	height uint16
	size   int64
}

func NewReader(inFile *os.File) *FlvReader {
	fi, _ := inFile.Stat()
	return &FlvReader{
		InFile: inFile,
		width:  0,
		height: 0,
		size:   fi.Size(),
	}
}

type FlvWriter struct {
	OutFile *os.File
}

func NewWriter(outFile *os.File) *FlvWriter {
	return &FlvWriter{
		OutFile: outFile,
	}
}

func (frReader *FlvReader) ReadHeader() (*Header, error) {
	header := make([]byte, HEADER_LENGTH+4)
	_, err := frReader.InFile.Read(header)
	if err != nil {
		return nil, err
	}

	sig := header[0:3]
	if bytes.Compare(sig, []byte(SIG)) != 0 {
		return nil, fmt.Errorf("bad file format")
	}
	version := (uint16(header[3]) << 8) | (uint16(header[4]) << 0)
	//skip := header[4:5]
	//offset := header[5:9]
	//next_id := header[9:13]

	return &Header{Version: version, Body: header}, nil
}

func (frWriter *FlvWriter) WriteHeader(header *Header) error {
	_, err := frWriter.OutFile.Write(header.Body)
	if err != nil {
		return err
	}
	return nil
}

func (frWriter *FlvWriter) WriteHeaderToStringIO(header *Header, sio *stringio.StringIO) error {
	_, err := sio.Write(header.Body)
	if err != nil {
		return err
	}
	return nil
}

func (fr *FlvReader) Recover(e Error, scanLength int) (broken Frame, err error, seekLength int) {
	re, ok := e.(*ReadError)
	if !ok {
		return nil, fmt.Errorf("unrecoverable read error"), 0
	}
	// fmt.Printf("\n%v %d\n", re, scanLength)

	scanStart := re.position
	readStart := re.position
	scanBuf := []byte{}

	if re.incomplete != nil {
		scanBuf = re.incomplete.Body
		readStart += int64(len(scanBuf))
		scanLength += len(scanBuf)
	}

	fr.InFile.Seek(readStart, os.SEEK_SET)
	b := make([]byte, scanLength)
	_, err = fr.InFile.Read(b)

	if err != nil {
		return nil, Unrecoverable(err.Error(), readStart), 0
	}

	scanBuf = append(scanBuf, b...)
	// fmt.Printf("%v\n", scanBuf)
	validTagStart := []byte{8, 9, 18}
	seekLength = 0
	for {
		for ; (seekLength < scanLength) && (bytes.IndexByte(validTagStart, scanBuf[seekLength]) == -1); seekLength++ {
		}
		if seekLength == scanLength {
			return nil, fmt.Errorf("no valid frames @[%d-%d]", scanStart, int(scanStart)+seekLength), seekLength
		}
		fr.InFile.Seek(scanStart+int64(seekLength), os.SEEK_SET)
		_, err := fr.readFrame()
		if err == nil {
			break
		}
		seekLength += 1
		if seekLength == scanLength {
			return nil, fmt.Errorf("no valid frames @[%d-%d]", scanStart, int(scanStart)+seekLength), seekLength
		}
	}
	fr.InFile.Seek(scanStart+int64(seekLength), os.SEEK_SET)

	if re.incomplete != nil {
		f := re.incomplete
		f.Body = f.Body[:seekLength]
		broken = fr.parseFrame(f)
	}

	return
}

func (frReader *FlvReader) readFrame() (*CFrame, Error) {
	var n int

	curPos, err := frReader.InFile.Seek(0, os.SEEK_CUR)
	if err != nil {
		return nil, Unrecoverable(err.Error(), curPos)
	}

	tagHeaderB := make([]byte, TAG_HEADER_LENGTH)
	n, err = frReader.InFile.Read(tagHeaderB)
	if n == 0 {
		return nil, nil
	}
	if TagSize(n) != TAG_HEADER_LENGTH {
		return nil, Unrecoverable(fmt.Sprintf("bad tag length=%d", n), curPos)
	}
	if err != nil {
		return nil, Unrecoverable(err.Error(), curPos)
	}

	validTagStart := []byte{8, 9, 18}
	if bytes.IndexByte(validTagStart, tagHeaderB[0]) == -1 {
		return nil, InvalidTagStart(curPos)
	}
	tagType := TagType(tagHeaderB[0])

	bodyLen := (uint32(tagHeaderB[1]) << 16) | (uint32(tagHeaderB[2]) << 8) | (uint32(tagHeaderB[3]) << 0)
	ts := (uint32(tagHeaderB[4]) << 16) | (uint32(tagHeaderB[5]) << 8) | (uint32(tagHeaderB[6]) << 0)
	tsExt := uint32(tagHeaderB[7])
	stream := (uint32(tagHeaderB[8]) << 16) | (uint32(tagHeaderB[9]) << 8) | (uint32(tagHeaderB[10]) << 0)

	var dts uint32
	dts = (tsExt << 24) | ts

	bodyBuf := make([]byte, bodyLen)
	n, err = frReader.InFile.Read(bodyBuf)
	if err != nil {
		return nil, Unrecoverable(err.Error(), curPos)
	}

	prevTagSizeB := make([]byte, PREV_TAG_SIZE_LENGTH)
	n, err = frReader.InFile.Read(prevTagSizeB)
	if err != nil {
		return nil, Unrecoverable(err.Error(), curPos)
	}
	prevTagSize := (uint32(prevTagSizeB[0]) << 24) | (uint32(prevTagSizeB[1]) << 16) | (uint32(prevTagSizeB[2]) << 8) | (uint32(prevTagSizeB[3]) << 0)

	pFrame := &CFrame{
		Stream:      stream,
		Dts:         dts,
		Type:        tagType,
		Position:    curPos,
		Body:        bodyBuf,
		PrevTagSize: prevTagSize,
	}
	if prevTagSize != bodyLen+uint32(TAG_HEADER_LENGTH) {
		return nil, IncompleteFrameError(pFrame)
	}
	return pFrame, nil
}

func (frReader *FlvReader) parseFrame(pFrame *CFrame) (resFrame Frame) {
	bodyBuf := pFrame.Body
	tagType := pFrame.Type

	switch tagType {
	case TAG_TYPE_META:
		pFrame.Flavor = METADATA
		resFrame = MetaFrame{CFrame: pFrame}
	case TAG_TYPE_VIDEO:
		if len(bodyBuf) > 0 {
			vft := VideoFrameType(uint8(bodyBuf[0]) >> 4)
			codecId := VideoCodec(uint8(bodyBuf[0]) & 0x0F)
			switch vft {
			case VIDEO_FRAME_TYPE_KEYFRAME:
				pFrame.Flavor = KEYFRAME
			default:
				pFrame.Flavor = FRAME
			}

			switch {
			case codecId == VIDEO_CODEC_ON2VP6 && vft == VIDEO_FRAME_TYPE_KEYFRAME:
				hHelper := (uint16(bodyBuf[1]) >> 4) & 0x0F
				wHelper := uint16(bodyBuf[1]) & 0x0F
				w := uint16(bodyBuf[5])
				h := uint16(bodyBuf[6])

				frReader.width = w*16 - wHelper
				frReader.height = h*16 - hHelper
				fmt.Printf("hHelper:%v, wHelper:%v, w:%v, h:%v\n", hHelper, wHelper, w, h)
			case codecId == VIDEO_CODEC_AVC && AvcPacketType(bodyBuf[1]) == VIDEO_AVC_SEQUENCE_HEADER:
				confRecord, err := ParseAVCConfRecord(bodyBuf[5:])
				if err == nil {
					// fmt.Printf("\nparsed %s\n", confRecord)
					sps, err := ParseSPS(confRecord.RawSPSData[0])
					if err == nil {
						// fmt.Printf("parsed %s\n\n", sps)
						frReader.width = uint16(sps.Width())
						frReader.height = uint16(sps.Height())
					} else {
						// fmt.Printf("\nparse error %s\n\n", err)
					}
					fmt.Printf("w:%v, h:%v\n", frReader.width, frReader.height)
				} else {
					// fmt.Printf("\nparse error %s\n\n", err)
				}
			}

			fmt.Printf("width:%v, height:%v\n", frReader.width, frReader.height)
			vFrame := VideoFrame{CFrame: pFrame, CodecId: codecId, Width: frReader.width, Height: frReader.height}
			switch codecId {
			case VIDEO_CODEC_AVC:
				resFrame = AVCVideoFrame{VideoFrame: &vFrame, PacketType: AvcPacketType(bodyBuf[1])}
			default:
				resFrame = vFrame
			}
		} else {
			resFrame = VideoFrame{CFrame: pFrame, CodecId: VIDEO_CODEC_UNDEFINED, Width: frReader.width, Height: frReader.height}
		}
	case TAG_TYPE_AUDIO:
		pFrame.Flavor = FRAME
		if len(bodyBuf) > 0 {
			codecId := AudioCodec(uint8(bodyBuf[0]) >> 4)
			rate := audioRate(AudioRate((uint8(bodyBuf[0]) >> 2) & 0x03))
			bitSize := AudioSize((uint8(bodyBuf[0]) >> 1) & 0x01)
			channels := AudioType(uint8(bodyBuf[0]) & 0x01)
			resFrame = AudioFrame{CFrame: pFrame, CodecId: codecId, Rate: rate, BitSize: bitSize, Channels: channels}
		} else {
			resFrame = AudioFrame{CFrame: pFrame, CodecId: AUDIO_CODEC_UNDEFINED, Rate: audioRate(AUDIO_RATE_UNDEFINED), BitSize: AUDIO_SIZE_UNDEFINED, Channels: AUDIO_TYPE_UNDEFINED}
		}
	}

	return resFrame
}

func (frReader *FlvReader) ReadFrame() (resFrame Frame, err Error) {
	pFrame, err := frReader.readFrame()
	if err != nil {
		return
	}
	if pFrame != nil {
		resFrame = frReader.parseFrame(pFrame)
	}
	return
}

func audioRate(ar AudioRate) uint32 {
	var ret uint32
	switch ar {
	case AUDIO_RATE_5_5:
		ret = 5500
	case AUDIO_RATE_11:
		ret = 11000
	case AUDIO_RATE_22:
		ret = 22000
	case AUDIO_RATE_44:
		ret = 44000
	default:
		ret = 0
	}
	return ret
}

func (frWriter *FlvWriter) WriteFrame(fr Frame) (e error) {
	return fr.WriteFrame(frWriter.OutFile)
}

func (frWriter *FlvWriter) WriteFrameToStringIO(fr Frame, sio *stringio.StringIO) (e error) {
	return fr.WriteFrame(sio)
}
