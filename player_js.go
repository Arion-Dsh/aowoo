//go:build js && !itest
// +build js,!itest

package aowoo

import (
	"io"
	"reflect"
	"runtime"
	"sync"
	"syscall/js"
	"unsafe"
)

var theJsp *jsplayer

type jsplayer struct {
	sources map[*Source]*jsSource
	ready   bool
	cond    *sync.Cond

	rate int
	chs  int

	ctx js.Value

	periodSize int // 32
}

func Open(sampleRate, bitsDeth, channels int) {
	var ctx js.Value
	ctx = js.Global().Get("AudioContext")
	if !ctx.Truthy() {
		ctx = js.Global().Get("webkitAudioContext")
	}
	if !ctx.Truthy() {
		panic("aowoo: js audio not support")
	}

	theJsp = &jsplayer{
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

		if !theJsp.ready {
			theJsp.ctx.Call("resume")
			theJsp.ready = true
			close(ready)
			theJsp.cond.Broadcast()
		}

		js.Global().Get("document").Call("removeEventListener", "keyup", cb)
		return nil
	})

	js.Global().Get("document").Call("addEventListener", "keyup", cb)

}

func NewSource(src io.ReadCloser, volmue float32, autoPlay bool) (*Source, error) {

	theJsp.cond.L.Lock()
	if !theJsp.ready {
		theJsp.cond.Wait()

	}
	theJsp.cond.L.Unlock()

	if volmue > 1 || volmue < 0 {
		return nil, ErrorSourceValumeRange
	}

	s := &Source{
		volmue: volmue,
	}

	b, _ := io.ReadAll(src)

	f32 := jsF32Array(b)
	l := f32.Get("length").Int() / theJsp.chs
	buff := theJsp.ctx.Call("createBuffer", theJsp.chs, l, theJsp.rate)

	for ch := 0; ch < theJsp.chs; ch++ {
		cbuff := buff.Call("getChannelData", ch)
		for i := 0; i < l; i++ {
			cbuff.SetIndex(i, f32.Index(ch+i*theJsp.chs))
		}
	}

	source := theJsp.ctx.Call("createBufferSource")
	source.Set("buffer", buff)
	source.Call("connect", theJsp.ctx.Get("destination"))
	if autoPlay {
		source.Call("start")
	}

	theJsp.sources[s] = &jsSource{bufferSource: source}
	return s, nil
}

func Play(ss ...*Source) {
	for _, s := range ss {
		buff, ok := theJsp.sources[s]
		if !ok {
			continue
		}
		buff.bufferSource.Call("start")
	}
}

func Stop(ss ...*Source) {

	for _, s := range ss {
		buff, ok := theJsp.sources[s]
		if ok {
			buff.bufferSource.Call("stop")
			delete(theJsp.sources, s)
		}
	}

}

func Puase(ss ...*Source)  {}
func Resume(ss ...*Source) {}

type jsSource struct {
	// buf          []byte
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
