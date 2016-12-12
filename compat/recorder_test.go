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

package compat

import (
	"errors"
	"reflect"
	"testing"

	"github.com/zagrodzki/goscope/scope"
)

func newRecorder() (*Recorder, chan<- []scope.ChannelData, func() []scope.Data) {
	r := &Recorder{
		TB: scope.Millisecond,
	}
	var got []scope.Data
	done := make(chan struct{})

	ch := make(chan []scope.ChannelData)
	r.Reset(scope.Microsecond, ch)
	d := r.Data

	go func() {
		for rcvd := range d {
			got = append(got, rcvd)
		}
		close(done)
	}()

	return r, ch, func() []scope.Data {
		<-done
		return got
	}
}

func TestRecorder(t *testing.T) {
	_, ch, done := newRecorder()
	ch <- []scope.ChannelData{
		{
			ID:      "one",
			Samples: []scope.Voltage{1, 2, 3},
		},
		{
			ID:      "two",
			Samples: []scope.Voltage{4, 5, 6},
		},
	}
	close(ch)
	got := done()
	want := []scope.Data{
		{
			Channels: []scope.ChannelData{
				{
					ID:      "one",
					Samples: []scope.Voltage{1, 2, 3},
				},
				{
					ID:      "two",
					Samples: []scope.Voltage{4, 5, 6},
				},
			},
			Num:      3,
			Interval: scope.Microsecond,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got data sequence %+v, want %+v", got, want)
	}
}

func TestRecorderError(t *testing.T) {
	r, ch, done := newRecorder()
	sampleErr := errors.New("foo")
	r.Error(sampleErr)
	close(ch)
	got := done()
	want := []scope.Data{
		{
			Error: sampleErr,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got data sequence %+v, want %+v", got, want)
	}
}
