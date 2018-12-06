package speexdsp

import (
	"math"
	"testing"
)

const (
	inLenGen = 960
	stereo   = 2
	mono     = 1
)

func TestInit(t *testing.T) {
	inLen, channels := inLenGen, 2
	pcm := makeSinePcm(inLen, channels)
	r, err := ResamplerInit(2, 48000, 48000, 4)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 1000; i++ {
		if _, _, err := r.PocessIntInterleaved(pcm); err != nil {
			t.Error(err)
		}
	}
	r.Destroy()
}
func TestError(t *testing.T) {
	errors := []string{
		"Success.",
		"Memory allocation failed.",
		"Bad resampler state.",
		"Invalid argument.",
		"Input and output buffers overlap.",
	}
	unknownError := "Unknown error. Bad error code or strange version mismatch."
	for i := 0; i < ErrorMaxError+100; i++ {
		s := StrError(i)
		if i >= ErrorMaxError {
			if s.Error() != unknownError {
				t.Error("error code mismatch text", i)
			}
		} else if s.Error() != errors[i] {
			t.Error("error code mismatch text", i)
		}
	}
}

// makes interleaved pcm odd channels=sin, even=cos
func makeSinePcm(samples, channels int) []int16 {
	pcm := make([]int16, samples*channels)
	for s := 0; s < samples; s++ {
		sin, cos := math.Sincos(math.Pi * 2 / float64(samples) * float64(s))
		sin *= math.MaxInt16
		cos *= math.MaxInt16
		for c := 0; c < channels; c++ {
			if c&1 == 0 {
				pcm[s*channels+c] = int16(sin)
			} else {
				pcm[s*channels+c] = int16(cos)
			}
		}
	}
	return pcm
}

func TestProcessInt(t *testing.T) {
	fromBase := int(48000)
	pcm := makeSinePcm(inLenGen, mono)
	r, err := ResamplerInit(mono, fromBase, fromBase, QualityDefault)
	if err != nil {
		t.Error(err)
	}
	if _, _, err := r.PocessInt(1, pcm); err == nil {
		t.Error("PocessInt returns noerr on errored channel")
	}
	x := [](*Resampler){}
	for i := 0.1; i < 2; i += .01 {
		toBase := int(float64(fromBase) * i)
		r, err := ResamplerInit(mono, fromBase, toBase, QualityDefault)
		if err != nil {
			t.Error(err)
		}
		x = append(x, r)
		pos := 0
		out := 0
		steps := 0
		// speexdsp is used as "Black Box", we dont know all situations, when
		// resampler returns earlier, than input ends
		for q := 1; q < 100; q++ {
			for pos < len(pcm) {
				readed, resPcm, err := r.PocessInt(0, pcm[pos:])
				if err != nil {
					t.Error(err)
					break
				}
				out += len(resPcm)
				pos += readed
				steps++
			}
			if math.Abs(float64(out)/float64(len(pcm))-i) > 1e-2 {
				t.Error(i, steps, inLenGen, out)
			}
		}
	}
	for _, r := range x {
		r.Destroy()
	}
}

func TestProcessIntInterleaved(t *testing.T) {
	fromBase := int(48000)
	inLen := inLenGen
	channels := stereo
	pcm := makeSinePcm(inLen, channels)
	var r *Resampler
	var err error
	var x [](*Resampler)
loop:
	for i := 0.1; i < 2; i += .01 {
		toBase := int(float64(fromBase) * i)
		r, err = ResamplerInit(channels, fromBase, toBase, QualityDefault)
		if err != nil {
			t.Error(err)
		}
		x = append(x, r)
		pos := 0
		out := 0
		steps := 0
		// speexdsp is used as "Black Box", we dont know all situations, when
		// resampler returns earlier, than input ends
		for pos < len(pcm) {
			readed, resPcm, err := r.PocessIntInterleaved(pcm[pos:])
			if err != nil {
				t.Error(err)
				break loop
			}
			out += len(resPcm)
			pos += readed
			steps++
		}
		if math.Abs(float64(out)/float64(len(pcm))-i) > 1e-2 {
			t.Error(i, steps, inLen, out)
			break loop
		}

	}
	for _, r := range x {
		r.Destroy()
	}
}

// makes interleaved float32 pcm odd channels=sin, even=cos
func makeSinePcmFloat32(samples, channels int) []float32 {
	pcm := make([]float32, samples*channels)
	for s := 0; s < samples; s++ {
		sin, cos := math.Sincos(math.Pi * 2 / float64(samples) * float64(s))
		for c := 0; c < channels; c++ {
			if c&1 == 0 {
				pcm[s*channels+c] = float32(sin)
			} else {
				pcm[s*channels+c] = float32(cos)
			}
		}
	}
	return pcm
}
func TestProcessFloat(t *testing.T) {
	fromBase := int(48000)
	inLen := inLenGen
	channels := mono
	pcm := makeSinePcmFloat32(inLen, channels)

	var r *Resampler
	var err error

	r, err = ResamplerInit(channels, fromBase, fromBase, QualityDefault)
	if _, _, err := r.PocessFloat(1, pcm); err == nil {
		t.Error("PocessFloat returns noerr on errored channel")
	}
	x := [](*Resampler){}
	for i := 0.5; i < 2; i += .01 {
		toBase := int(float64(fromBase) * i)
		r, err = ResamplerInit(channels, fromBase, toBase, QualityDefault)
		if err != nil {
			t.Error(err)
		}
		x = append(x, r)
		pos := 0
		out := 0
		steps := 0
		// speexdsp is used as "Black Box", we dont know all situations, when
		// resampler returns earlier, than input ends
		for pos < len(pcm) {
			readed, resPcm, err := r.PocessFloat(0, pcm[pos:])
			if err != nil {
				t.Error(err)
				break
			}
			out += len(resPcm)
			pos += readed
			steps++
		}
		if math.Abs(float64(out)/float64(len(pcm))-i) > 0.1 {
			t.Error(i, steps, inLen, out, float64(toBase)/float64(fromBase), float64(out)/float64(len(pcm)))
		}
	}
	for _, r := range x {
		r.Destroy()
	}
}

func TestProcessFloatInterleaved(t *testing.T) {
	fromBase := int(48000)
	inLen := inLenGen
	channels := 2
	pcm := makeSinePcmFloat32(inLen, channels)
	x := [](*Resampler){}
loop:
	for i := 0.1; i < 2; i += .01 {
		toBase := int(float64(fromBase) * i)
		r, err := ResamplerInit(channels, fromBase, toBase, QualityDefault)
		if err != nil {
			t.Error(err)
		}
		x = append(x, r)
		pos := 0
		out := 0
		steps := 0
		// speexdsp is used as "Black Box", we dont know all situations, when
		// resampler returns earlier, than input ends
		for pos < len(pcm) {
			readed, resPcm, err := r.PocessFloatInterleaved(pcm[pos:])
			if err != nil {
				t.Error(err)
				break loop
			}
			out += len(resPcm)
			pos += readed
			steps++
		}
		if math.Abs(float64(out)/float64(len(pcm))-i) > 1e-2 {
			t.Error(i, steps, inLen, out)
			break loop
		}

	}
	for _, r := range x {
		r.Destroy()
	}
}

func TestFrac(t *testing.T) {
	inLen, channels := inLenGen, 2
	pcm := makeSinePcm(inLen, channels)
	num, denum := 7, 11
	r, err := ResamplerInitFrac(2, num, denum, 48000, 48000, 4)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 10; i++ {
		if _, _, err := r.PocessIntInterleaved(pcm); err != nil {
			t.Error(err)
		}
	}
	if num1, denum1, err := r.GetRatio(); err == nil {
		if num != num1 || denum != denum1 {
			t.Error("Ratio error")
		}
	} else {
		t.Error(err)
	}
	newNum, newDenum := 17, 13
	if err := r.SetRateFrac(newNum, newDenum, 48000, 48000); err != nil {
		t.Error(err)
	}

	if num1, denum1, err := r.GetRatio(); err == nil {
		if newNum != num1 || newDenum != denum1 {
			t.Error("Ratio error")
		}
	} else {
		t.Error(err)
	}
	r.Destroy()
}

func TestLatency(t *testing.T) {
	inLen, channels := inLenGen, 2
	pcm := makeSinePcm(inLen, channels)
	inF, outF := 48000, 44100
	r, err := ResamplerInit(2, inF, outF, 4)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 10; i++ {
		if _, _, err := r.PocessIntInterleaved(pcm); err != nil {
			t.Error(err)
		}
	}
	if latency, err := r.GetOutputLatency(); err != nil {
		t.Error(err)
	} else {
		if latency == 0 {
			t.Error("Latency error")
		}
	}
	if latency, err := r.GetInputLatency(); err != nil {
		t.Error(err)
	} else {
		if latency == 0 {
			t.Error("Latency error")
		}
	}
	r.Destroy()
}

func TestRate(t *testing.T) {
	inLen, channels := inLenGen, 2
	pcm := makeSinePcm(inLen, channels)
	inF, outF := 48000, 44100
	r, err := ResamplerInit(2, inF, outF, 4)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 10; i++ {
		if _, _, err := r.PocessIntInterleaved(pcm); err != nil {
			t.Error(err)
		}
	}
	if inF1, outF1, err := r.GetRate(); err == nil {
		if inF != inF1 || outF != outF1 {
			t.Error("Rate error")
		}
	} else {
		t.Error(err)
	}
	inF, outF = outF, inF
	if err := r.SetRate(inF, outF); err != nil {
		t.Error(err)
	}
	if inF1, outF1, err := r.GetRate(); err == nil {
		if inF != inF1 || outF != outF1 {
			t.Error("Rate error")
		}
	} else {
		t.Error(err)
	}

	r.Destroy()
}

func TestQuality(t *testing.T) {
	inF, outF := 48000, 44100
	r, err := ResamplerInit(2, inF, outF, 4)
	if err != nil {
		t.Error(err)
	}

	for i := QualityMin; i <= QualityMax; i++ {
		if err := r.SetQuality(i); err != nil {
			t.Error(err)
			break
		}
	}
	for _, i := range []int{QualityMin - 1, QualityMax + 1} {
		if err := r.SetQuality(i); err == nil {
			t.Error("SetQuality return noerr w/quality ", i)
			break
		}
	}

	r.Destroy()
}

func TestSkipZeros(t *testing.T) {
	inF, outF := 48000, 44100
	r, err := ResamplerInit(2, inF, outF, 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.SkipZeros(); err != nil {
		t.Error(err)
	}
	r.Destroy()
}

func TestResetMem(t *testing.T) {
	inF, outF := 48000, 44100
	r, err := ResamplerInit(2, inF, outF, 4)
	if err != nil {
		t.Error(err)
	}

	if err := r.ResetMem(); err != nil {
		t.Error(err)
	}
	r.Destroy()
}
