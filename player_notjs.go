//go:build !js
// +build !js

package aowoo

import (
	"io"
	"sync"
)

func Open(sampleRate, bitsdepth, channels int) {
	if p == nil {
		p = &player{
			cond:    sync.NewCond(new(sync.Mutex)),
			state:   playing,
			sources: map[io.ReadCloser]*Source{},
		}
	}
	p.set(sampleRate, bitsdepth, channels)
	p.driver = newDriver(sampleRate, bitsdepth, channels)
	p.driver.setCallback(p.callback)
}

func NewSource(src io.ReadCloser, volmue float32, autoPlay bool) (*Source, error) {

	if p == nil {
		return nil, ErrorAowooNotOpen
	}

	if s, ok := p.sources[src]; ok {
		return s, ErrorSourceAlreadyLoaded
	}

	if volmue > 1 || volmue < 0 {
		return nil, ErrorSourceValumeRange
	}

	s := &Source{
		volmue: volmue,
		src:    src,
		pause:  true,
	}

	if autoPlay {
		s.pause = false
		p.sources[src] = s
	}
	return s, nil
}

func Play(ss ...*Source) {
	for _, s := range ss {
		p.sources[s.src] = s
		go func(sr *Source) {
			p.cond.L.Lock()
			defer p.cond.L.Unlock()
			p.cond.Wait()
			sr.pause = false
		}(s)
	}

	p.cond.Broadcast()

}

func Stop(ss ...*Source) {
	p.delSource(ss...)
}

func Resume(ss ...*Source) {
	for _, s := range ss {
		s.pause = false
	}
}

func Puase(ss ...*Source) {
	for _, s := range ss {
		s.pause = true
	}

}

func (p *player) delSource(ss ...*Source) {
	for _, s := range ss {
		s.pause = true
		s.src.Close()
		delete(p.sources, s.src)
	}
}

func (p *player) callback(b []float32) (int, error) {

	var done []*Source
	read := 0
	for _, s := range p.sources {
		if s.pause {
			continue
		}
		d := make([]byte, len(b)*2)
		n, _ := s.src.Read(d)
		if n == 0 {
			done = append(done, s)
		}
		if n > read {
			read = n
		}
		for i := 0; i < len(b); i++ {
			b[i] = bf32(d[2*i:]) * s.volmue
		}

	}

	p.delSource(done...)

	return read, nil
}
