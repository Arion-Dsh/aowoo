package aowoo

import (
	"errors"
	"io"
	"sync"
)

var ErrorSourceAlreadyLoaded = errors.New("error: source already in aowoo.")
var ErrorSourceValumeRange = errors.New("aowoo source must betwee 0 and 1")
var ErrorAowooNotOpen = errors.New("Aowoo must be open before!")

var p *player

type player struct {
	mu       sync.Mutex
	cond     *sync.Cond
	sources  map[io.ReadCloser]*Source
	rate     int
	depth    int
	channels int

	driver driver

	state state
}

func (p *player) set(sampleRate, bitsdepth, channels int) {
	p.rate = sampleRate
	p.depth = bitsdepth
	p.channels = channels

}

type Source struct {
	mu  sync.Mutex
	src io.ReadCloser

	volmue float32

	pause bool
}

// byte to float32 / 1<< 15
func bf32(b []byte) float32 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	i := int16(b[0]) | int16(b[1])<<8
	return float32(i) / (1 << 15)
}
