//go:build !js
// +build !js

package github.com/arion-dsh/aowoo

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
	p.cond.Broadcast()
}

func NewSource(src io.ReadCloser, valmue float32, autoPlay bool) *Source {
	if s, ok := p.sources[src]; ok {
		return s
	}

	if valmue > 1 || valmue < 0 {
		panic("aowoo source must betwee 0 and 1")
	}
	s := &Source{
		id:     newAowooID(),
		valmue: valmue,
		src:    src,
		pause:  true,
	}
	if autoPlay {
		s.pause = false
		p.sources[src] = s
	}
	return s
}

func PlaySource(ss ...*Source) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	if p.state != playing {
		p.cond.Wait()
	}
	for _, s := range ss {
		if s == nil {
			panic("aowoo Source not be nil")
		}
		s.pause = false
		p.sources[s.src] = s
	}
}

func DelSource(ss ...*Source) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	p.delSource(ss...)
}
func (p *player) Resume() {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	p.state = playing
	p.cond.Broadcast()
}

func (p *player) Puase() {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	p.state = paused
	p.cond.Broadcast()

}
func (p *player) delSource(ss ...*Source) {
	for _, s := range ss {
		s.src.Close()
		delete(p.sources, s.src)
	}
}

func (p *player) callback(b []float32) (int, error) {

	if p.state != playing {
		return 0, nil
	}

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
			b[i] += bf32(d[2*i:]) * s.valmue
		}

	}

	p.delSource(done...)

	return read, nil
}
