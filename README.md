## speexdsp lib PCM audio Resampler

[![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/aam335/speexdsp) [![Build](https://github.com/aam335/speexdsp/actions/workflows/build.yml/badge.svg)](https://github.com/aam335/speexdsp/actions/workflows/build.yml)

`speexdsp` is a Golang bindings for [libspeexdsp](https://gitlab.xiph.org/xiph/speexdsp), provides PCM sample rate converter for PCM coding.

### Installation

Linux:

```
sudo apt-get install libspeexdsp-dev
go get -v github.com/aam335/speexdsp
```

macOS:

```
brew install speexdsp
go get -v github.com/aam335/speexdsp
```

### Example

```go
r, err := speexdsp.ResamplerInit(2, 48000, 44100, speexdsp.QualityDefault)
if err != nil {
	panic(err)
}
defer r.Destroy()

var inpcm []int16
// fill "inpcm"

if readed, outpcm, err := r.PocessIntInterleaved(inpcm); err != nil {
	panic(err)
}

// "readed" contains number of int16 values read from "inpcm"/
// "outpcm" contains resampling result (length may differ from "inpcm")
```

You can find more examples in [tests](speexdsp_test.go).
