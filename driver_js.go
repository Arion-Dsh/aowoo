//go:build js && !itest
// +build js,!itest

package github.com/arion-dsh/aowoo

import (
	"io"
	"reflect"
	"runtime"
	"sync"
	"syscall/js"
	"unsafe"
)

type jsdriver struct {
	sources map[*Source]*jsSource
	ready   bool
	cond    *sync.Cond

	rate int
	chs  int

	ctx js.Value

	periodSize int // 32
}

func NewPlayer(sampleRate, bitsDeth, channels int) Player {
	var ctx js.Value
	ctx = js.Global().Get("AudioContext")
	if !ctx.Truthy() {
		ctx = js.Global().Get("webkitAudioContext")
	}
	if !ctx.Truthy() {
		panic("aowoo: js audio not support")
	}

	d := &jsdriver{
		sources: map[*Source]*jsSource{},
		cond:    sync.NewCond(new(sync.Mutex)),
		ready:   false,

		rate: sampleRate,
		chs:  channels,

		ctx:        ctx.New(),
		periodSize: 32,
	}
	ready := make(chan struct{})
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, arguments []js.Value) interface{} {

		if !d.ready {
			d.ctx.Call("resume")
			d.ready = true
			close(ready)
			d.cond.Broadcast()
		}

		js.Global().Get("document").Call("removeEventListener", "keyup", cb)
		return nil
	})

	js.Global().Get("document").Call("addEventListener", "keyup", cb)

	return d

}

func (d *jsdriver) NewSource(r io.Reader, valmue float32, autoPlay bool) *Source {
	// go func() {
	d.cond.L.Lock()
	if !d.ready {
		d.cond.Wait()
	}
	d.cond.L.Unlock()

	b, _ := io.ReadAll(r)
	f32 := jsF32Array(b)
	l := f32.Get("length").Int() / d.chs
	buff := d.ctx.Call("createBuffer", d.chs, l, d.rate)

	for ch := 0; ch < d.chs; ch++ {
		cbuff := buff.Call("getChannelData", ch)
		for i := 0; i < l; i++ {
			cbuff.SetIndex(i, f32.Index(ch+i*d.chs))
		}
	}

	source := d.ctx.Call("createBufferSource")
	source.Set("buffer", buff)
	source.Call("connect", d.ctx.Get("destination"))
	if autoPlay {
		source.Call("start")
	}

	// }()
	return nil
}

func (d *jsdriver) PlaySource(ss ...*Source) {
	d.cond.L.Lock()
	defer d.cond.L.Unlock()
	for _, s := range ss {
		buff, ok := d.sources[s]
		if !ok {
			continue
		}
		buff.bufferSource.Call("start")
	}
}

func (d *jsdriver) DelSource(ss ...*Source) {
	d.cond.L.Lock()
	defer d.cond.L.Unlock()
	for _, s := range ss {
		buff, ok := d.sources[s]
		if ok {
			buff.bufferSource.Call("stop")
			delete(d.sources, s)
		}
	}

}

func (d *jsdriver) Puase()  {}
func (d *jsdriver) Resume() {}

type jsSource struct {
	s *Source

	buf          []byte
	bufferSource js.Value
}

func jsF32Array(b []byte) js.Value {
	var f32 []float32
	for i := 0; i < len(b)/2; i++ {
		v := bf32([]byte{b[i*2], b[i*2+1]})
		f32 = append(f32, v)
	}

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&f32))
	sh.Len *= 4
	sh.Cap *= 4
	bs := *(*[]byte)(unsafe.Pointer(sh))
	runtime.KeepAlive(b)

	u := js.Global().Get("Uint8Array").New(len(bs))
	js.CopyBytesToJS(u, bs)
	buf := u.Get("buffer")
	offset := u.Get("byteOffset")
	l := u.Get("byteLength").Int() / 4
	return js.Global().Get("Float32Array").New(buf, offset, l)

}
