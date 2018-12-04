package speexdsp

/*
#cgo pkg-config: speexdsp
#include <speex/speex_resampler.h>

*/
import "C"
import (
	"errors"
	"log"
	"unsafe"
)

// Resampler ...
type Resampler struct {
	resampler                          (*C.SpeexResamplerState)
	outBuff                            []int16 // one of this buffers used when typed data readed
	outBuffFloat                       []float32
	channels, inRate, outRate, quality int
	multiplier                         float32
	in, out                            unsafe.Pointer
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

var v = 0

// ResamplerInit init new remuxer
func ResamplerInit(channels, inRate, outRate, quality int) (*Resampler, error) {
	log.Print("+1 ", v)
	v++
	err := C.int(0)
	r := &Resampler{channels: channels, inRate: inRate, outRate: outRate, quality: quality}
	r.multiplier = float32(outRate) / float32(inRate) * 1.1 // 10% перестраховка ;)
	r.resampler = C.speex_resampler_init(C.spx_uint32_t(channels),
		C.spx_uint32_t(inRate),
		C.spx_uint32_t(outRate),
		C.int(quality),
		&err)
	if r.resampler == nil {
		return r, StrError(int(err))
	}
	return r, nil
}

// Destroy resampler
func (r *Resampler) Destroy() error {
	if r.resampler != nil {
		C.speex_resampler_destroy(r.resampler)
		return nil
	}
	return StrError(ErrorInvalidArg)
}

// PocessInt Resample an int slice
func (r *Resampler) PocessInt(channel int, in []int16) (int, []int16, error) {
	outBuffCap := int(float32(len(in)) * r.multiplier)
	if outBuffCap > cap(r.outBuff) {
		r.outBuff = make([]int16, int(float32(outBuffCap)*reserve))
	}
	inLen := C.spx_uint32_t(len(in))
	outLen := C.spx_uint32_t(outBuffCap)
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
	outBuffCap := int(float32(len(in))*r.multiplier) + 1000
	// if outBuffCap > cap(r.outBuff) {
	// 	r.outBuff = make([]int16, int(float32(outBuffCap)*reserve))
	// }
	inn := make([]C.spx_int16_t, len(in))
	out := make([]C.spx_int16_t, outBuffCap)
	// inLen := C.spx_uint32_t(len(in))
	// outLen := C.spx_uint32_t(outBuffCap)
	x := []C.spx_uint32_t{0, 0}
	x[0] = C.spx_uint32_t(len(in))
	x[1] = C.spx_uint32_t(outBuffCap)
	res := C.speex_resampler_process_interleaved_int(
		r.resampler,
		&inn[0],
		&x[0], //inLen,
		&out[0],
		&x[1], //outLen,
	)
	if res != ErrorSuccess {
		return 0, nil, StrError(ErrorInvalidArg)
	}
	//	return int(inLen), r.outBuff[:outLen], nil
	return int(x[0]), make([]int16, x[1]), nil
}

// PocessFloat Resample an float32 slice
func (r *Resampler) PocessFloat(channel int, in []float32) (int, []float32, error) {
	outBuffCap := int(float32(len(in)) * r.multiplier)
	if outBuffCap > cap(r.outBuffFloat) {
		r.outBuffFloat = make([]float32, int(float32(outBuffCap)*reserve))
	}
	inLen := C.spx_uint32_t(len(in))
	outLen := C.spx_uint32_t(outBuffCap)
	res := C.speex_resampler_process_float(
		r.resampler,
		C.spx_uint32_t(channel),
		(*C.float)(&in[0]),
		&inLen,
		(*C.float)(&r.outBuffFloat[0]),
		&outLen,
	)
	if res != ErrorSuccess {
		return 0, nil, StrError(ErrorInvalidArg)
	}
	return int(inLen), r.outBuffFloat[:outLen], nil
}

// PocessFloatInterleaved Resample an int slice interleaved
func (r *Resampler) PocessFloatInterleaved(in []float32) (int, []float32, error) {
	outBuffCap := int(float32(len(in)) * r.multiplier)
	if outBuffCap > cap(r.outBuffFloat) {
		r.outBuffFloat = make([]float32, int(float32(outBuffCap)*reserve))
	}
	inLen := C.spx_uint32_t(len(in))
	outLen := C.spx_uint32_t(outBuffCap)
	res := C.speex_resampler_process_interleaved_float(
		r.resampler,
		(*C.float)(&in[0]),
		&inLen,
		(*C.float)(&r.outBuffFloat[0]),
		&outLen,
	)
	if res != ErrorSuccess {
		return 0, nil, StrError(ErrorInvalidArg)
	}
	return int(inLen), r.outBuffFloat[:outLen], nil
}

// StrError returns Returns the English meaning for an error code
func StrError(errorCode int) error {
	cS := C.speex_resampler_strerror(C.int(errorCode))
	if cS == nil {
		return nil
	}
	return errors.New(C.GoString(cS))
}
