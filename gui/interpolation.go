//  Copyright 2016 The goscope Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package gui

import (
	"fmt"

	"github.com/mjibson/go-dsp/fft"
	"github.com/zagrodzki/goscope/scope"
)

type interpMethod interface {
	interpolate(samples []scope.Sample, size int) ([]scope.Sample, error)
}

type constInterpolation struct{}

func (interp constInterpolation) interpolate(samples []scope.Sample, size int) ([]scope.Sample, error) {
	if err := checkSizes(len(samples), size); err != nil {
		return nil, err
	}
	if len(samples) == 1 {
		return constFunction(size, samples[0]), nil
	}

	interpSamples := make([]scope.Sample, size)
	interval := float64(size-1) / float64(len(samples)-1)
	lastIndex := 0
	lastInterp := 0.0
	nextInterp := interval
	for i, _ := range interpSamples {
		fi := float64(i)
		if fi-lastInterp > nextInterp-fi {
			lastIndex++
			lastInterp = float64(lastIndex) * interval
			nextInterp = float64(lastIndex+1) * interval
		}
		interpSamples[i] = samples[lastIndex]
	}
	return interpSamples, nil
}

type linearInterpolation struct {
}

func (interp linearInterpolation) interpolate(samples []scope.Sample, size int) ([]scope.Sample, error) {
	if err := checkSizes(len(samples), size); err != nil {
		return nil, err
	}
	if len(samples) == 1 {
		return constFunction(size, samples[0]), nil
	}

	interpSamples := make([]scope.Sample, size)
	interval := float64(size-1) / float64(len(samples)-1)
	lastIndex := 0
	lastInterp := 0.0
	nextInterp := interval
	a := float64(samples[lastIndex+1]-samples[lastIndex]) / (nextInterp - lastInterp)
	b := float64(samples[lastIndex]) - a*lastInterp
	for i, _ := range interpSamples {
		if float64(i) > nextInterp {
			lastIndex++
			lastInterp = float64(lastIndex) * interval
			nextInterp = float64(lastIndex+1) * interval
			a = float64(samples[lastIndex+1]-samples[lastIndex]) / (nextInterp - lastInterp)
			b = float64(samples[lastIndex]) - a*lastInterp
		}
		interpSamples[i] = scope.Sample(a*float64(i) + b)
	}
	return interpSamples, nil
}

type sincInterpolation struct {
}

func (interp sincInterpolation) interpolate(samples []scope.Sample, size int) ([]scope.Sample, error) {
	if err := checkSizes(len(samples), size); err != nil {
		return nil, err
	}

	floatSamples := make([]float64, len(samples))
	for i, s := range samples {
		floatSamples[i] = float64(s)
	}
	frequencies := fft.FFTReal(floatSamples)
	freqInterp := make([]complex128, size)
	freqSize := len(frequencies)
	cmplxFreqSize := complex(float64(freqSize), 0)
	for i := 0; i < freqSize/2; i++ {
		freqInterp[i] = frequencies[i] / cmplxFreqSize
		freqInterp[len(freqInterp)-1-i] = frequencies[freqSize-1-i] / cmplxFreqSize
	}
	cmplxInterpSamples := fft.IFFT(freqInterp)

	interpSamples := make([]scope.Sample, size)
	floatSize := float64(size)
	for i, s := range cmplxInterpSamples {
		interpSamples[i] = scope.Sample(floatSize * real(s))
	}
	return interpSamples, nil
}

func checkSizes(samplesSize, requestedSize int) error {
	if samplesSize == 0 {
		return fmt.Errorf("input samples should contain at least one element")
	}
	if samplesSize >= requestedSize {
		return fmt.Errorf("requested samples size (%v) is less or equal than input samples size (%v)", requestedSize, samplesSize)
	}
	return nil
}

func constFunction(size int, value scope.Sample) []scope.Sample {
	samples := make([]scope.Sample, size)
	for i, _ := range samples {
		samples[i] = value
	}
	return samples
}
