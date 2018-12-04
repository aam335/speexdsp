package speexdsp

import (
	"log"
	"math"
	"runtime"
	"testing"
)

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	log.Print("Alloc = ", bToKb(m.Alloc), " TotalAlloc = ", bToKb(m.TotalAlloc), " Sys = ", bToKb(m.Sys), " NumGC = ", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
func bToKb(b uint64) uint64 {
	return b / 1024
}

const (
	inLenGen = 960
	stereo   = 2
	mono     = 1
)

func TestInit(t *testing.T) {
	// x := [](*Resampler){}
	// cnt := 200
	inLen, channels := 960, 2
	pcm := makeSinePcm(inLen, channels)
	//runtime.GC()
	// for i := 0; i < cnt; i++ {
	//	PrintMemUsage()
	r, err := ResamplerInit(2, 48000, 44100, 4)
	// x = append(x, r)
	if err != nil {
		t.Error(err)
	}
	// PrintMemUsage()
	//	PrintMemUsage()
	// log.Print(i, unsafe.Pointer(r.resampler))
	// }
	// log.Print(len(x))
	/*
		for i, r := range x {
			log.Print(i, unsafe.Pointer(r.resampler))
			//_, _, _ :=
		r.PocessIntInterleaved(pcm)
			// rpcm = nil
		}
	*/
	for i := 0; i < 1000; i++ {
		r.PocessIntInterleaved(pcm)
	}
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
	// inLen := 960
	// channels := 1
	pcm := makeSinePcm(inLenGen, mono)
	for i := 0.1; i < 2; i += .05 {
		toBase := int(float64(fromBase) * i)
		r, err := ResamplerInit(mono, fromBase, toBase, QualityDefault)
		if err != nil {
			t.Error(err)
		}
		pos := 0
		out := 0
		steps := 0
		// speexdsp is used as "Black Box", we dont know all situations, when
		// resampler returns earlier, than input ends
		for q := 1; q < 100; q++ {
			for pos < len(pcm) {
				readed, resPcm, err := r.PocessInt(1, pcm[pos:])
				if err != nil {
					t.Error(err)
				}
				out += len(resPcm)
				pos += readed
				steps++
			}
			if math.Abs(float64(out)/float64(len(pcm))-i) > 1e-5 {
				t.Error(i, steps, inLenGen, out)
			}
		}
	}
}

func TestProcessIntInterleaved(t *testing.T) {
	fromBase := int(48000)
	inLen := inLenGen
	channels := stereo
	pcm := makeSinePcm(inLen, channels)
	for i := 0.1; i < 2; i += .05 {
		toBase := int(float64(fromBase) * i)
		r, err := ResamplerInit(channels, fromBase, toBase, QualityDefault)
		if err != nil {
			t.Error(err)
		}
		pos := 0
		out := 0
		steps := 0
		// speexdsp is used as "Black Box", we dont know all situations, when
		// resampler returns earlier, than input ends
		for pos < len(pcm) {
			readed, resPcm, err := r.PocessIntInterleaved(pcm[pos:])
			if err != nil {
				t.Error(err)
			}
			out += len(resPcm)
			pos += readed
			steps++
		}
		if math.Abs(float64(out)/float64(len(pcm))-i) > 1e-5 {
			t.Error(i, steps, inLen, out)
		}

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
func TestProcessFloat32(t *testing.T) {
	fromBase := int(48000)
	inLen := inLenGen
	channels := mono
	pcm := makeSinePcmFloat32(inLen, channels)
	for i := 0.5; i < 2; i += .01 {
		toBase := int(float64(fromBase) * i)
		r, err := ResamplerInit(channels, fromBase, toBase, QualityDefault)
		if err != nil {
			t.Error(err)
		}
		pos := 0
		out := 0
		steps := 0
		// speexdsp is used as "Black Box", we dont know all situations, when
		// resampler returns earlier, than input ends
		for pos < len(pcm) {
			readed, resPcm, err := r.PocessFloat(1, pcm[pos:])
			if err != nil {
				t.Error(err)
			}
			out += len(resPcm)
			pos += readed
			steps++
		}
		if math.Abs(float64(out)/float64(len(pcm))-i) > 1e-5 {
			t.Error(i, steps, inLen, out)
		}

	}
}

func TestProcessFloatInterleaved(t *testing.T) {
	fromBase := int(48000)
	inLen := 200
	channels := 2
	pcm := makeSinePcmFloat32(inLen, channels)
	for i := 0.5; i < 2; i += .1 {
		toBase := int(float64(fromBase) * i)
		r, err := ResamplerInit(channels, fromBase, toBase, QualityDefault)
		if err != nil {
			t.Error(err)
		}
		pos := 0
		out := 0
		steps := 0
		// speexdsp is used as "Black Box", we dont know all situations, when
		// resampler returns earlier, than input ends
		for pos < len(pcm) {
			readed, resPcm, err := r.PocessFloatInterleaved(pcm[pos:])
			if err != nil {
				t.Error(err)
			}
			out += len(resPcm)
			pos += readed
			steps++
		}
		if math.Abs(float64(out)/float64(len(pcm))-i) > 1e-5 {
			t.Error(i, steps, inLen, out)
		}

	}
}

func TestTTT(t *testing.T) {
	shit()

}
