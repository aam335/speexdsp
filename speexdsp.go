package speexdsp

/*
#cgo pkg-config: speexdsp
#include <speex/speex_resampler.h>
*/
import "C"
import (
	"errors"
)

// Resampler ...
type Resampler struct {
	resampler                          *C.SpeexResamplerState
	outBuff                            []int16 // one of this buffers used when typed data readed
	outBuffFloat                       []float32
	channels, inRate, outRate, quality int
	multiplier                         float32
}

// Quality
const (
	QualityMax     = 10
	QualityMin     = 0
	QualityDefault = 4
	QualityDesktop = 5
	QualityVoid    = 3
)

// Errors
const (
	ErrorSuccess = iota
	ErrorAllocFailed
	ErrorBadState
	ErrorInvalidArg
	ErrorPtrOverlap
	ErrorMaxError
)

// some constants
const (
	reserve = 1.1
)

// ResamplerInit init new remuxer
func ResamplerInit(channels, inRate, outRate, quality int) (*Resampler, error) {
	err := C.int(0)
	r := &Resampler{channels: channels, inRate: inRate, outRate: outRate, quality: quality}
	r.multiplier = float32(outRate) / float32(inRate) * 1.1 // 10% перестраховка ;)
	r.resampler = C.speex_resampler_init(C.spx_uint32_t(channels),
		C.spx_uint32_t(inRate), C.spx_uint32_t(outRate), C.int(quality), &err)
	return r, nil
}

// Destroy resampler
func (r *Resampler) Destroy() error {
	if r.resampler != nil {
		C.speex_resampler_destroy((*C.SpeexResamplerState)(r.resampler))
		return nil
	}
	return StrError(ErrorInvalidArg)
}

// PocessInt Resample an int slice
func (r *Resampler) PocessInt(channel int, in []int16) (int, []int16, error) {
	if channel >= r.channels {
		return 0, nil, StrError(ErrorInvalidArg)
	}
	if int(float32(len(in))*r.multiplier) > len(r.outBuff) {
		r.outBuff = make([]int16, int(float32(len(in))*r.multiplier*reserve))
	}
	inLen := C.spx_uint32_t(len(in))
	outLen := C.spx_uint32_t(len(r.outBuff))
	res := C.speex_resampler_process_int(
		r.resampler,
		C.spx_uint32_t(channel),
		(*C.spx_int16_t)(&in[0]),
		&inLen,
		(*C.spx_int16_t)(&r.outBuff[0]),
		&outLen,
	)
	if res != ErrorSuccess {
		return 0, nil, StrError(ErrorInvalidArg)
	}
	return int(inLen), r.outBuff[:outLen], nil
}

// PocessIntInterleaved Resample an int slice interleaved
func (r *Resampler) PocessIntInterleaved(in []int16) (int, []int16, error) {
	outBuffCap := int(float32(len(in)) * r.multiplier)
	if outBuffCap > cap(r.outBuff) {
		r.outBuff = make([]int16, int(float32(outBuffCap)*reserve)*4)
	}
	inLen := C.spx_uint32_t(len(in) / r.channels)
	outLen := C.spx_uint32_t(len(r.outBuff) / r.channels)
	res := C.speex_resampler_process_interleaved_int(
		r.resampler,
		(*C.spx_int16_t)(&in[0]),
		&inLen,
		(*C.spx_int16_t)(&r.outBuff[0]),
		&outLen,
	)
	if res != ErrorSuccess {
		return 0, nil, StrError(ErrorInvalidArg)
	}
	return int(inLen) * r.channels, r.outBuff[:outLen*2], nil
}

// PocessFloat Resample an float32 slice
func (r *Resampler) PocessFloat(channel int, in []float32) (int, []float32, error) {
	if channel >= r.channels {
		return 0, nil, StrError(ErrorInvalidArg)
	}
	if int(float32(len(in))*r.multiplier) > len(r.outBuffFloat) {
		r.outBuffFloat = make([]float32, int(float32(len(in))*r.multiplier*reserve))
	}
	inLen := C.spx_uint32_t(len(in))
	outLen := C.spx_uint32_t(len(r.outBuffFloat))
	res := C.speex_resampler_process_float(
		r.resampler,
		C.spx_uint32_t(channel),
		(*C.float)(&in[0]),
		&inLen,
		(*C.float)(&r.outBuffFloat[0]),
		&outLen,
	)
	if res != ErrorSuccess {
		return 0, nil, StrError(int(res))
	}
	return int(inLen), r.outBuffFloat[:outLen], nil
}

// PocessFloatInterleaved Resample an int slice interleaved
func (r *Resampler) PocessFloatInterleaved(in []float32) (int, []float32, error) {
	if int(float32(len(in))*r.multiplier) > len(r.outBuffFloat) {
		r.outBuffFloat = make([]float32, int(float32(len(in))*r.multiplier*reserve))
	}
	inLen := C.spx_uint32_t(len(in) / r.channels)
	outLen := C.spx_uint32_t(len(r.outBuffFloat) / r.channels)
	res := C.speex_resampler_process_interleaved_float(
		(*C.SpeexResamplerState)(r.resampler),
		(*C.float)(&in[0]),
		&inLen,
		(*C.float)(&r.outBuffFloat[0]),
		&outLen,
	)
	if res != ErrorSuccess {
		return 0, nil, StrError(ErrorInvalidArg)
	}
	return int(inLen) * r.channels, r.outBuffFloat[:int(outLen)*r.channels], nil
}

// StrError returns Returns the English meaning for an error code
func StrError(errorCode int) error {
	cS := C.speex_resampler_strerror(C.int(errorCode))
	if cS == nil {
		return nil
	}
	return errors.New(C.GoString(cS))
}
