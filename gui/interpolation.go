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

// Interpolator constructs new data points within the range of a set of known data points
type Interpolator func([]scope.Voltage, int) ([]scope.Voltage, error)

// StepInterpolator assigns the value of the nearest data point
// performing piecewise constant interpolation
func StepInterpolator(samples []scope.Voltage, size int) ([]scope.Voltage, error) {

	if err := checkSizes(len(samples), size); err != nil {
		return nil, err
	}
	if len(samples) == 1 {
		return constValue(size, samples[0]), nil
	}

	interpSamples := make([]scope.Voltage, size)
	interval := float64(size-1) / float64(len(samples)-1)
	lastIndex := 0
	lastInterp := 0.0
	nextInterp := interval
	for i := range interpSamples {
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

// LinearInterpolator estimates the values of the points
// using linear segments joining neihbouring points
func LinearInterpolator(samples []scope.Voltage, size int) ([]scope.Voltage, error) {

	if err := checkSizes(len(samples), size); err != nil {
		return nil, err
	}
	if len(samples) == 1 {
		return constValue(size, samples[0]), nil
	}

	interpSamples := make([]scope.Voltage, size)
	interval := float64(size-1) / float64(len(samples)-1)
	lastIndex := 0
	lastInterp := 0.0
	nextInterp := interval
	a := float64(samples[lastIndex+1]-samples[lastIndex]) / (nextInterp - lastInterp)
	b := float64(samples[lastIndex]) - a*lastInterp
	for i := range interpSamples {
		if float64(i) > nextInterp {
			lastIndex++
			lastInterp = float64(lastIndex) * interval
			nextInterp = float64(lastIndex+1) * interval
			a = float64(samples[lastIndex+1]-samples[lastIndex]) / (nextInterp - lastInterp)
			b = float64(samples[lastIndex]) - a*lastInterp
		}
		interpSamples[i] = scope.Voltage(a*float64(i) + b)
	}
	return interpSamples, nil
}

// SincInterpolator uses Fourier series for interpolation
func SincInterpolator(samples []scope.Voltage, size int) ([]scope.Voltage, error) {
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

	interpSamples := make([]scope.Voltage, size)
	floatSize := float64(size)
	for i, s := range cmplxInterpSamples {
		interpSamples[i] = scope.Voltage(floatSize * real(s))
	}
	return interpSamples, nil
}

// SincZeroPadInterpolator uses Fourier series for interpolation and adds
// zeros to the input samples to make sure their number is a power of 2
func SincZeroPadInterpolator(samples []scope.Voltage, size int) ([]scope.Voltage, error) {
	samplesLen := len(samples)
	padSamplesLen := samplesLen
	mask := 1 << 20
	for mask != 0 {
		if mask&samplesLen != 0 {
			if mask == samplesLen {
				padSamplesLen = mask
			} else {
				padSamplesLen = mask << 1
			}
			break
		}
		mask = mask >> 1
	}
	if padSamplesLen == samplesLen {
		return SincInterpolator(samples, size)
	}
	padLenLeft := (padSamplesLen - len(samples)) / 2
	padLenRight := padSamplesLen - len(samples) - padLenLeft
	padSamples := append(make([]scope.Voltage, padLenLeft), samples...)
	padSamples = append(padSamples, make([]scope.Voltage, padLenRight)...)
	padInterpSize := round(float64(padSamplesLen*size) / float64(len(samples)))
	interpolated, err := SincInterpolator(padSamples, padInterpSize)
	if err != nil {
		return nil, err
	}
	return interpolated[(padInterpSize-size)/2 : (padInterpSize+size)/2], nil
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

func constValue(size int, value scope.Voltage) []scope.Voltage {
	samples := make([]scope.Voltage, size)
	for i := range samples {
		samples[i] = value
	}
	return samples
}
