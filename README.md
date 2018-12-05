## speexdsp
`speexdsp` is a Golang bindings for libspeexdsp, provides PCM sample rate converter for PCM coding

### Installation 

sudo apt-get install libspeexdsp-dev
go get -v github.com/aam335/speexdsp

### Examples
```go
...
	r, err := speexdsp.ResamplerInit(2, 48000, 44100, 4)
	if err != nil {
		t.Error(err)
	}
    for _,pcm: = range pcms { // pcms=[][]int16  
        // pcm =[]int16 
		if readed, outpcm, err := r.PocessIntInterleaved(pcm); err != nil {
            t.Error(err)
            break
		}
    }
    r.Destroy()
...
```
for more info see tests

