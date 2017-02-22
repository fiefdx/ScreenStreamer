package flv

import (
    "fmt"
    "bytes"
)

type AVCProfile byte
const (
    AVC_PROFILE_BASELINE   AVCProfile = 66
    AVC_PROFILE_MAIN       AVCProfile = 77
    AVC_PROFILE_EXTENDED   AVCProfile = 88
    AVC_PROFILE_HIGH       AVCProfile = 100
    AVC_PROFILE_HIGH10     AVCProfile = 110
    AVC_PROFILE_HIGH422    AVCProfile = 122
    AVC_PROFILE_HIGH444    AVCProfile = 244
    AVC_PROFILE_CAVLC444   AVCProfile = 44
)

var (
    avcProfileStrings = map[AVCProfile]string {
        AVC_PROFILE_BASELINE:   "Baseline",
        AVC_PROFILE_MAIN:       "Main",
        AVC_PROFILE_EXTENDED:   "Extended",
        AVC_PROFILE_HIGH:       "High",
        AVC_PROFILE_HIGH10:     "High 10",
        AVC_PROFILE_HIGH422:    "High 4:2:2",
        AVC_PROFILE_HIGH444:    "High 4:4:4",
        AVC_PROFILE_CAVLC444:   "CAVLC 4:4:4",
    }
)

func (p AVCProfile) String() string {
    return avcProfileStrings[p]
}


type AVCConfRecord struct {
    ConfigurationVersion byte
    AVCProfileIndication AVCProfile
    ProfileCompatibility byte
    AVCLevelIndication  byte
    RawSPSData [][]byte
    RawPPSData [][]byte
}

func (r *AVCConfRecord) String() string {
    return fmt.Sprintf("AVCConfigurationRecord(ver. %d, profile: %s, level: %d, %d SPS, %d PPS)",
        r.ConfigurationVersion, r.AVCProfileIndication,
        r.AVCLevelIndication,
        len(r.RawSPSData), len(r.RawPPSData))
}

func ParseAVCConfRecord(data []byte) (rec *AVCConfRecord, err error) {
    r := NewBitReader(data)

    defer func () {
        if rec := recover(); rec != nil {
            err = rec.(error)
        }
    }()

    configurationVersion := r.U8()
    AVCProfileIndication := r.U8()
    profile_compatibility := r.U8()
    AVCLevelIndication := r.U8()

    r.U(6)
    /* nobody follows standard in reserved
    if r.U(6) != 077 {
        panic("wrong reserved 1")
    } */

    r.U(2) /* lengthSizeMinusOne */
    r.U(3)
    /* same here
    if r.U(3) != 07 {
        panic("wrong reserved 2")
    } */

    numOfSPS := r.U(5)
    spss := make([][]byte, numOfSPS)
    for i := uint32(0); i < numOfSPS; i++ {
        spsLen := r.U(16)
        spss[i] = make([]byte, spsLen)
        r.Read(spss[i])
    }

    numOfPPS := r.U(8)
    ppss := make([][]byte, numOfPPS)
    for i := uint32(0); i < numOfPPS; i++ {
        ppsLen := r.U(16)
        ppss[i] = make([]byte, ppsLen)
        r.Read(ppss[i])
    }

    rec = &AVCConfRecord{
            ConfigurationVersion: configurationVersion,
            AVCProfileIndication: AVCProfile(AVCProfileIndication),
            ProfileCompatibility: profile_compatibility,
            AVCLevelIndication: AVCLevelIndication,
            RawSPSData: spss,
            RawPPSData: ppss,
        }
    return
}


type SPS struct {
    Profile_idc AVCProfile
    Constraint_set byte
    Level_idc byte
    SPS_id uint32

    pic_width_in_mbs uint32
    pic_height_in_map_units uint32
    frame_mbs_only_flag uint32
    crops FrameCropOffsets
}

type FrameCropOffsets struct {
    left uint32
    right uint32
    top uint32
    bottom uint32
}

func (sps *SPS) Width() uint32 {
    w := sps.pic_width_in_mbs*16 - sps.crops.left*2 - sps.crops.right*2
    return w
}

func (sps *SPS) Height() uint32 {
    c := uint32(2) - sps.frame_mbs_only_flag
    h := sps.pic_height_in_map_units*16 - sps.crops.top*2 - sps.crops.bottom*2
    return c*h
}

func (sps *SPS) String() string {
    return fmt.Sprintf("seq_parameter_set(profile: %d, level: %d, id: %d)", sps.Profile_idc, sps.Level_idc, sps.SPS_id)
}

func ParseSPS(rawSPSNALU []byte) (ret *SPS, err error) {
    r := NewBitReader(rawSPSNALU)

    defer func () {
        if rec := recover(); rec != nil {
            err = rec.(error)
        }
    }()

    r.U(1) /* forbidden_zero_bit */
    r.U(2) /* nal_ref_idc */

    nal_unit_type := r.U(5)
    if nal_unit_type != 7 {
        err = fmt.Errorf("Not SPS NALU, nal_unit_type = %d", nal_unit_type)
        return
    }

    profile_idc := r.U8()

    constraint_set_flags := byte(0)
    for i := uint(6); i > 0; i-- {
        f := byte(r.U(1))
        constraint_set_flags |= f << (i - 1)
    }

    r.U(2) /* reserved_zero_2bits */
    level_idc := r.U8()

    seq_parameter_set_id := r.Ue()

    extended_profiles := []byte{100, 110, 122, 244, 44, 83, 86, 118, 128}
    if bytes.IndexByte(extended_profiles, profile_idc) != -1 {

        chroma_format_idc := r.Ue()
        if chroma_format_idc == 3 {
            r.U(1) // separate_colour_plane_flag
        }
        r.Ue() // bit_depth_luma_minus8
        r.Ue() // bit_depth_chroma_minus8
        r.U(1) // qpprime_y_zero_transform_bypass_flag
        seq_scaling_matrix_present_flag := r.U(1)
        if seq_scaling_matrix_present_flag != 0 {
            c := 12
            if chroma_format_idc != 3 {
                c = 8
            }

            for i := 0; i < c; i++ {
                seq_scaling_list_present_flag := r.U(1)
                if seq_scaling_list_present_flag != 0 {
                    if i < 6 {
                        scaling_list(r, 16)
                    } else {
                        scaling_list(r, 64)
                    }
                }
            }
        }
    }

    r.Ue() /* log2_max_frame_num_minus4 */
    pic_order_cnt_type := r.Ue()
    if pic_order_cnt_type == 0 {
        r.Ue() /* log2_max_pic_order_cnt_lsb_minus4 */
    } else if pic_order_cnt_type == 1 {
        r.U(1) /* delta_pic_order_always_zero_flag */
        r.Se() /* offset_for_non_ref_pic */
        r.Se() /* offset_for_top_to_bottom_field */
        num_ref_frames_in_pic_order_cnt_cycle := r.Ue()
        for i := uint32(0); i <num_ref_frames_in_pic_order_cnt_cycle; i++ {
            r.Se() /* offset_for_ref_frame[ i ] */
        }
    }

    r.Ue() /* max_num_ref_frames */
    r.U(1) /* gaps_in_frame_num_value_allowed_flag */
    pic_width_in_mbs_minus1 := r.Ue()
    pic_height_in_map_units_minus1 := r.Ue()

    frame_mbs_only_flag := r.U(1)
    if frame_mbs_only_flag == 0 {
        r.U(1) /* mb_adaptive_frame_field_flag */
    }

    r.U(1) /* direct_8x8_inference_flag */

    crops := FrameCropOffsets{}

    frame_cropping_flag := r.U(1)
    if frame_cropping_flag != 0 {
        frame_crop_left_offset := r.Ue()
        frame_crop_right_offset := r.Ue()
        frame_crop_top_offset := r.Ue()
        frame_crop_bottom_offset := r.Ue()

        crops = FrameCropOffsets{
                    left: frame_crop_left_offset,
                    right: frame_crop_right_offset,
                    top: frame_crop_top_offset,
                    bottom: frame_crop_bottom_offset,
                }
    }

    ret = &SPS{
                Profile_idc: AVCProfile(profile_idc),
                Constraint_set: constraint_set_flags,
                Level_idc: level_idc,
                SPS_id: seq_parameter_set_id,

                pic_width_in_mbs: pic_width_in_mbs_minus1 + 1,
                pic_height_in_map_units : pic_height_in_map_units_minus1 + 1,
                frame_mbs_only_flag : frame_mbs_only_flag,
                crops: crops,
            }
    return
}

func scaling_list(r *BitReader, scalingListSize uint32) {
    lastScale := int32(8)
    nextScale := int32(8)

    for j := uint32(0); j < scalingListSize; j++ {
        if nextScale != 0 {
            delta_scale := r.Se()
            nextScale = (lastScale + delta_scale + 256) % 256
        }
        if nextScale != 0 {
            lastScale = nextScale
        }
    }
}
