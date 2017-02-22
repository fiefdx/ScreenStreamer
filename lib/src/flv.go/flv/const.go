package flv

const (
	SIG = "FLV"
)

type TagSize byte

const (
	HEADER_LENGTH        TagSize = 9
	PREV_TAG_SIZE_LENGTH TagSize = 4
	TAG_HEADER_LENGTH    TagSize = 11
)

type TagType byte

const (
	TAG_TYPE_AUDIO TagType = 8
	TAG_TYPE_VIDEO TagType = 9
	TAG_TYPE_META  TagType = 18
)

type VideoFrameType byte

const (
	VIDEO_FRAME_TYPE_KEYFRAME         VideoFrameType = 1
	VIDEO_FRAME_TYPE_INTER_FRAME      VideoFrameType = 2
	VIDEO_FRAME_TYPE_DISP_INTER_FRAME VideoFrameType = 3
	VIDEO_FRAME_TYPE_GENERATED        VideoFrameType = 4
	VIDEO_FRAME_TYPE_COMMAND          VideoFrameType = 5
	VIDEO_FRAME_TYPE_UNDEFINED        VideoFrameType = 255
)

type VideoCodec byte

const (
	VIDEO_CODEC_JPEG         VideoCodec = 1
	VIDEO_CODEC_SORENSON     VideoCodec = 2
	VIDEO_CODEC_SCREENVIDEO  VideoCodec = 3
	VIDEO_CODEC_ON2VP6       VideoCodec = 4
	VIDEO_CODEC_ON2VP6_ALPHA VideoCodec = 5
	VIDEO_CODEC_SCREENVIDEO2 VideoCodec = 6
	VIDEO_CODEC_AVC          VideoCodec = 7
	VIDEO_CODEC_H264         VideoCodec = 28
	VIDEO_CODEC_UNDEFINED    VideoCodec = 255
)

type AvcPacketType byte

const (
	VIDEO_AVC_SEQUENCE_HEADER AvcPacketType = 0
	VIDEO_AVC_NALU            AvcPacketType = 1
	VIDEO_AVC_SEQUENCE_END    AvcPacketType = 2
)

type AudioType byte

const (
	AUDIO_TYPE_MONO      AudioType = 0
	AUDIO_TYPE_STEREO    AudioType = 1
	AUDIO_TYPE_UNDEFINED AudioType = 255
)

type AudioSize byte

const (
	AUDIO_SIZE_8BIT      AudioSize = 0
	AUDIO_SIZE_16BIT     AudioSize = 1
	AUDIO_SIZE_UNDEFINED AudioSize = 255
)

type AudioRate byte

const (
	AUDIO_RATE_5_5       AudioRate = 0
	AUDIO_RATE_11        AudioRate = 1
	AUDIO_RATE_22        AudioRate = 2
	AUDIO_RATE_44        AudioRate = 3
	AUDIO_RATE_UNDEFINED AudioRate = 255
)

type AudioCodec byte

const (
	AUDIO_CODEC_PCM         AudioCodec = 0
	AUDIO_CODEC_ADPCM       AudioCodec = 1
	AUDIO_CODEC_MP3         AudioCodec = 2
	AUDIO_CODEC_PCM_LE      AudioCodec = 3
	AUDIO_CODEC_NELLYMOSER8 AudioCodec = 5
	AUDIO_CODEC_NELLYMOSER  AudioCodec = 6
	AUDIO_CODEC_A_G711      AudioCodec = 7
	AUDIO_CODEC_MU_G711     AudioCodec = 8
	AUDIO_CODEC_RESERVED    AudioCodec = 9
	AUDIO_CODEC_AAC         AudioCodec = 10
	AUDIO_CODEC_SPEEX       AudioCodec = 11
	AUDIO_CODEC_MP3_8KHZ    AudioCodec = 14
	AUDIO_CODEC_DEVICE      AudioCodec = 15
	AUDIO_CODEC_UNDEFINED   AudioCodec = 255
)

type AudioAac byte

const (
	AUDIO_AAC_SEQUENCE_HEADER AudioAac = 0
	AUDIO_AAC_RAW             AudioAac = 1
)

type Flavor byte

const (
	METADATA Flavor = iota
	FRAME
	KEYFRAME
)

var (
	vcToStr = map[VideoCodec]string{
		VIDEO_CODEC_JPEG:         "jpeg",
		VIDEO_CODEC_SORENSON:     "sorenson",
		VIDEO_CODEC_SCREENVIDEO:  "screen",
		VIDEO_CODEC_ON2VP6:       "vp6",
		VIDEO_CODEC_ON2VP6_ALPHA: "vp6a",
		VIDEO_CODEC_SCREENVIDEO2: "screen2",
		VIDEO_CODEC_AVC:          "avc",
	}

	ttToStr = map[TagType]string{
		TAG_TYPE_AUDIO: "audio",
		TAG_TYPE_VIDEO: "video",
		TAG_TYPE_META:  "meta",
	}

	vftToStr = map[VideoFrameType]string{
		VIDEO_FRAME_TYPE_KEYFRAME:         "keyframe",
		VIDEO_FRAME_TYPE_INTER_FRAME:      "frame",
		VIDEO_FRAME_TYPE_DISP_INTER_FRAME: "iframe",
		VIDEO_FRAME_TYPE_GENERATED:        "generated",
		VIDEO_FRAME_TYPE_COMMAND:          "command",
	}

	atToStr = map[AudioType]string{
		AUDIO_TYPE_MONO:   "mono",
		AUDIO_TYPE_STEREO: "stereo",
	}

	asToStr = map[AudioSize]string{
		AUDIO_SIZE_8BIT:  "8bit",
		AUDIO_SIZE_16BIT: "16bit",
	}

	arToStr = map[AudioRate]string{
		AUDIO_RATE_5_5: "5.5",
		AUDIO_RATE_11:  "11",
		AUDIO_RATE_22:  "22",
		AUDIO_RATE_44:  "44",
	}

	acToStr = map[AudioCodec]string{
		AUDIO_CODEC_PCM:         "pcm",
		AUDIO_CODEC_ADPCM:       "adpcm",
		AUDIO_CODEC_MP3:         "mp3",
		AUDIO_CODEC_PCM_LE:      "pcmle",
		AUDIO_CODEC_NELLYMOSER8: "nellymoser8",
		AUDIO_CODEC_NELLYMOSER:  "nellymoser",
		AUDIO_CODEC_A_G711:      "g711a",
		AUDIO_CODEC_MU_G711:     "g711u",
		AUDIO_CODEC_RESERVED:    "RESERVED",
		AUDIO_CODEC_AAC:         "aac",
		AUDIO_CODEC_SPEEX:       "speex",
		AUDIO_CODEC_MP3_8KHZ:    "mp3_8khz",
		AUDIO_CODEC_DEVICE:      "device",
	}

	avcptToStr = map[AvcPacketType]string{
		VIDEO_AVC_SEQUENCE_HEADER:	"sequence header",
		VIDEO_AVC_NALU:				"NALU",
		VIDEO_AVC_SEQUENCE_END:		"sequence end",
	}
)

func (vc VideoCodec) String() (s string) {
	return vcToStr[vc]
}

func (tt TagType) String() (s string) {
	return ttToStr[tt]
}

func (vft VideoFrameType) String() (s string) {
	return vftToStr[vft]
}

func (apt AvcPacketType) String() (s string) {
	return avcptToStr[apt]
}

func (at AudioType) String() (s string) {
	return atToStr[at]
}

func (as AudioSize) String() (s string) {
	return asToStr[as]
}

func (ar AudioRate) String() (s string) {
	return arToStr[ar]
}

func (ac AudioCodec) String() (s string) {
	return acToStr[ac]
}
