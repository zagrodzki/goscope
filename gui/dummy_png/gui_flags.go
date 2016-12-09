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

package main

import (
	"errors"
	"flag"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/zagrodzki/goscope/scope"
)

// yParams represents a custom command line flag for trace position parameters.
// Multiple flags in the format "chanID:zero,perDiv" can be passed to main(),
// each specifying the position of zero and volts per div for a given channel.
// For channels without custom parameters set in the flag default values are used.
type yParams map[scope.ChanID]scope.TraceParams

func (i *yParams) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *yParams) Set(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return errors.New("use format: \"chanID:zero,perDiv\"")
	}
	numbers := strings.Split(parts[1], ",")
	if len(numbers) != 2 {
		return errors.New("use format: \"chanID:zero,perDiv\"")
	}
	zero, err := strconv.ParseFloat(numbers[0], 64)
	if err != nil {
		return err
	}
	perDiv, err := strconv.ParseFloat(numbers[1], 64)
	if err != nil {
		return err
	}
	(*i)[scope.ChanID(parts[0])] = scope.TraceParams{zero, perDiv}
	return nil
}

// colParams represents a custom command line flag for color parameters.
// Multiple flags in the format "chanID:R,G,B" can be passed to main(),
// each specifying the color of a given channel's plot. For channels
// without custom parameters set in the flag default values are used.
type colParams map[scope.ChanID]color.RGBA

func (i *colParams) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *colParams) Set(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return errors.New("use format: \"chanID:R,G,B\"")
	}
	numbers := strings.Split(parts[1], ",")
	if len(numbers) != 3 {
		return errors.New("use format: \"chanID:R,G,B\"")
	}
	r, err := strconv.ParseUint(numbers[0], 10, 8)
	if err != nil {
		return err
	}
	g, err := strconv.ParseUint(numbers[1], 10, 8)
	if err != nil {
		return err
	}
	b, err := strconv.ParseUint(numbers[2], 10, 8)
	if err != nil {
		return err
	}
	(*i)[scope.ChanID(parts[0])] = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	return nil
}

func posFlag(flagName, flagHelp string) *yParams {
	tracePos := yParams{}
	flag.Var(&tracePos, flagName, flagHelp)
	return &tracePos
}

func colFlag(flagName, flagHelp string) *colParams {
	cols := colParams{}
	flag.Var(&cols, flagName, flagHelp)
	return &cols
}
