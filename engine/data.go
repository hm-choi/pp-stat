package engine

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

type HEData struct {
	ciphertexts []*rlwe.Ciphertext
	size        int
	level       int
	scale       float64
}

func (d *HEData) Size() int                       { return d.size }
func (d *HEData) Level() int                      { return d.level }
func (d *HEData) Scale() float64                  { return d.scale }
func (d *HEData) Ciphertexts() []*rlwe.Ciphertext { return d.ciphertexts }

func NewHEData(ciphertexts []*rlwe.Ciphertext, size int, level int, scale float64) *HEData {
	return &HEData{
		ciphertexts: ciphertexts,
		size:        size,
		level:       level,
		scale:       scale,
	}
}

func (d *HEData) CopyData() (cpData *HEData) {
	size := d.Size()
	level := d.Level()
	scale := d.Scale()

	ctxts := d.Ciphertexts()
	cpCtxts := make([]*rlwe.Ciphertext, len(ctxts))

	for i := 0; i < len(ctxts); i++ {
		cpCtxts[i] = ctxts[i].CopyNew()
	}

	return NewHEData(cpCtxts, size, level, scale)
}

func (d *HEData) Print() {
	fmt.Printf("[HEData] size=%d, level=%d, scale=%.3e\n", d.size, d.level, d.scale)
}
