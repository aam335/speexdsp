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
	resampler    *C.SpeexResamplerState
	outBuff      []int16 // one of this buffers used when typed data readed
	outBuffFloat []float32
	channels     int
	multiplier   float32
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

// ResamplerInit Create a new resampler with integer input and output rates
// Resampling quality between 0 and 10, where 0 has poor quality
// and 10 has very high quality
func ResamplerInit(channels, inRate, outRate, quality int) (*Resampler, error) {
	err := C.int(0)
	r := &Resampler{channels: channels}
	r.multiplier = float32(outRate) / float32(inRate) * 1.1 // 10% перестраховка ;)
	r.resampler = C.speex_resampler_init(C.spx_uint32_t(channels),
		C.spx_uint32_t(inRate), C.spx_uint32_t(outRate), C.int(quality), &err)
	if r.resampler == nil {
		return nil, StrError(int(err))
	}
	return r, nil
}

// ResamplerInitFrac Create a new resampler with fractional input/output rates. The sampling
// rate ratio is an arbitrary rational number with both the numerator and
// denominator being 32-bit integers.
func ResamplerInitFrac(channels, ratioNumerator, ratioDenuminator, inRate, outRate, quality int) (*Resampler, error) {
	err := C.int(0)
	r := &Resampler{channels: channels}
	r.multiplier = float32(outRate) / float32(inRate) * 1.1 // 10% перестраховка ;)
	r.resampler = C.speex_resampler_init_frac(C.spx_uint32_t(channels),
		C.spx_uint32_t(ratioNumerator), C.spx_uint32_t(ratioDenuminator),
		C.spx_uint32_t(inRate), C.spx_uint32_t(outRate), C.int(quality), &err)
	if r.resampler == nil {
		return nil, StrError(int(err))
	}
	return r, nil
}

// Destroy a resampler
func (r *Resampler) Destroy() error {
	if r.resampler != nil {
		C.speex_resampler_destroy((*C.SpeexResamplerState)(r.resampler))
		return nil
	}
	return StrError(ErrorInvalidArg)
}

// PocessInt Resample an int slice
// channel - Index of the channel to process for the multi-channel
// base (0 otherwise)
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
// channel - Index of the channel to process for the multi-channel
// base (0 otherwise)
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
		r.resampler,
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

// SetRate Set (change) the input/output sampling rates
func (r *Resampler) SetRate(inRate, outRate int) error {
	res := C.speex_resampler_set_rate(
		r.resampler,
		C.spx_uint32_t(inRate),
		C.spx_uint32_t(outRate),
	)
	if res != ErrorSuccess {
		return StrError(int(res))
	}
	return nil
}

// SetRateFrac Set (change) the input/output sampling rates and resampling ratio
// (fractional values in Hz supported).
func (r *Resampler) SetRateFrac(ratioNumerator, ratioDenuminator, inRate, outRate int) error {
	res := C.speex_resampler_set_rate_frac(
		r.resampler,
		C.spx_uint32_t(ratioNumerator),
		C.spx_uint32_t(ratioDenuminator),
		C.spx_uint32_t(inRate),
		C.spx_uint32_t(outRate),
	)
	if res != ErrorSuccess {
		return StrError(int(res))
	}
	return nil
}

// GetRate Get the current input/output sampling rates
func (r *Resampler) GetRate() (inRate int, outRate int, err error) {
	if r.resampler == nil {
		return 0, 0, StrError(ErrorInvalidArg)
	}
	var inR, outR C.spx_uint32_t
	C.speex_resampler_get_rate(
		r.resampler,
		&inR,
		&outR,
	)
	return int(inR), int(outR), nil
}

// GetRatio Get the current input/output sampling rates
func (r *Resampler) GetRatio() (ratioNumerator int, ratioDenuminator int, err error) {
	if r.resampler == nil {
		return 0, 0, StrError(ErrorInvalidArg)
	}
	var rNum, rDenum C.spx_uint32_t
	C.speex_resampler_get_ratio(
		r.resampler,
		&rNum,
		&rDenum,
	)
	return int(rNum), int(rDenum), nil
}

// GetQuality Get the conversion quality.
func (r *Resampler) GetQuality() (quality int, err error) {
	if r.resampler == nil {
		return 0, StrError(ErrorInvalidArg)
	}
	var q C.int
	C.speex_resampler_get_quality(
		r.resampler,
		&q,
	)
	return int(q), nil
}

// SetQuality Set (change) the conversion quality.
func (r *Resampler) SetQuality(quality int) error {
	if r.resampler == nil {
		return StrError(ErrorInvalidArg)
	}
	res := C.speex_resampler_set_quality(
		r.resampler,
		C.int(quality),
	)
	if res != ErrorSuccess {
		return StrError(int(res))
	}
	return nil
}

// GetInputStride Get the input stride.
func (r *Resampler) GetInputStride() (stride int, err error) {
	if r.resampler == nil {
		return 0, StrError(ErrorInvalidArg)
	}
	var s C.spx_uint32_t
	C.speex_resampler_get_input_stride(
		r.resampler,
		&s,
	)
	return int(s), nil
}

// SetInputStride Set (change) the input stride.
func (r *Resampler) SetInputStride(stride int) error {
	if r.resampler == nil {
		return StrError(ErrorInvalidArg)
	}
	C.speex_resampler_set_input_stride(
		r.resampler,
		C.spx_uint32_t(stride),
	)
	return nil
}

// GetOutputStride Get the output stride.
func (r *Resampler) GetOutputStride() (stride int, err error) {
	if r.resampler == nil {
		return 0, StrError(ErrorInvalidArg)
	}
	var s C.spx_uint32_t
	C.speex_resampler_get_output_stride(
		r.resampler,
		&s,
	)
	return int(s), nil
}

// SetOutputStride Set (change) the output stride.
func (r *Resampler) SetOutputStride(stride int) error {
	if r.resampler == nil {
		return StrError(ErrorInvalidArg)
	}
	C.speex_resampler_set_output_stride(
		r.resampler,
		C.spx_uint32_t(stride),
	)
	return nil
}

// GetInputLatency Get the input latency.
func (r *Resampler) GetInputLatency() (stride int, err error) {
	if r.resampler == nil {
		return 0, StrError(ErrorInvalidArg)
	}
	l := C.speex_resampler_get_input_latency(
		r.resampler,
	)
	return int(l), nil
}

// GetOutputLatency Get the output latency.
func (r *Resampler) GetOutputLatency() (stride int, err error) {
	if r.resampler == nil {
		return 0, StrError(ErrorInvalidArg)
	}
	l := C.speex_resampler_get_output_latency(
		r.resampler,
	)
	return int(l), nil
}

// SkipZeros Make sure that the first samples to go out of the resamplers don't have
// leading zeros. This is only useful before starting to use a newly created
// resampler. It is recommended to use that when resampling an audio file, as
// it will generate a file with the same length. For real-time processing,
// it is probably easier not to use this call (so that the output duration
// is the same for the first frame).
func (r *Resampler) SkipZeros() (err error) {
	if r.resampler == nil {
		return StrError(ErrorInvalidArg)
	}
	C.speex_resampler_skip_zeros(
		r.resampler,
	)
	return nil
}

// ResetMem Reset a resampler so a new (unrelated) stream can be processed
func (r *Resampler) ResetMem() (err error) {
	if r.resampler == nil {
		return StrError(ErrorInvalidArg)
	}
	C.speex_resampler_skip_zeros(
		r.resampler,
	)
	return nil
}

// StrError returns Returns the English meaning for an error code
func StrError(errorCode int) error {
	cS := C.speex_resampler_strerror(C.int(errorCode))
	if cS == nil {
		return nil
	}
	return errors.New(C.GoString(cS))
}
