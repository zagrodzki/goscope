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
	"log"
	"strconv"
	"strings"

	"github.com/zagrodzki/goscope/dummy"
	"github.com/zagrodzki/goscope/gui"
	"github.com/zagrodzki/goscope/scope"
)

type yParams map[scope.ChanID]gui.ZeroAndScale

func (i *yParams) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *yParams) Set(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return errors.New("use format: \"chanID:zero,scale\"")
	}
	numbers := strings.Split(parts[1], ",")
	if len(numbers) != 2 {
		return errors.New("use format: \"chanID:zero,scale\"")
	}
	zero, err := strconv.ParseFloat(numbers[0], 64)
	if err != nil {
		return err
	}
	scale, err := strconv.ParseFloat(numbers[1], 64)
	if err != nil {
		return err
	}
	(*i)[scope.ChanID(parts[0])] = gui.ZeroAndScale{zero, scale}
	return nil
}

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

func main() {
	dev, err := dummy.Open("")
	if err != nil {
		log.Fatalf("Cannot open the device: %v", err)
	}

	fileName := flag.String("file", "draw.png", "output file name")
	width := flag.Int("width", 800, "PNG width")
	height := flag.Int("height", 600, "PNG width")
	zas := yParams{}
	flag.Var(&zas, "zas", "zero and scale, format: \"chanID:zero,scale\"")
	cols := colParams{}
	flag.Var(&cols, "col", "color, format: \"chanID:R,G,B\"")
	flag.Parse()

	err = gui.PlotToPng(dev, *width, *height, zas, cols, *fileName)
	if err != nil {
		log.Fatalf("Cannot plot to file: %v", err)
	}
}
