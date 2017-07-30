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

package hantek6022be

import (
	"bufio"
	"log"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/shibukawa/configdir"
	"github.com/zagrodzki/goscope/scope"
)

type calVoltage struct {
	min      scope.Voltage
	interval scope.Voltage
}

type calData struct {
	voltage [2]map[rangeID]calVoltage
	time    map[rateID]scope.Duration
}

const calDataFile = "calibration.txt"

// parse and set the voltage calibration.
func parseVoltageCal(parts []string, cal *calData) error {
	if len(parts) != 5 {
		return errors.Errorf("voltage calibration should be volt:<channel_num>:<range_id>:<min_voltage>:<interval>, got only %d colon-separated parts", len(parts))
	}
	var ch int
	switch parts[1] {
	case "1":
		ch = 0
	case "2":
		ch = 1
	default:
		return errors.Errorf("invalid channel id %q, need 1 or 2", parts[1])
	}
	rng, err := strconv.ParseInt(parts[2], 16, 8)
	if err != nil {
		return errors.Errorf("invalid range ID %q, need a hexadecimal byte", parts[2])
	}
	if got := rangeIDToVolts[rangeID(rng)]; got == 0 {
		var ranges []string
		for k := range rangeIDToVolts {
			ranges = append(ranges, strconv.FormatInt(int64(k), 16))
		}
		return errors.Errorf("invalid range ID %q, want one of %v", parts[1], ranges)
	}
	min, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return errors.Errorf("invalid min voltage %q, need a float", parts[3])
	}
	interval, err := strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return errors.Errorf("invalid voltage interval %q, need a float", parts[4])
	}
	cal.voltage[ch][rangeID(rng)] = calVoltage{scope.Voltage(min), scope.Voltage(interval)}
	return nil
}

// parse and set the time calibration
func parseTimeCal(parts []string, cal *calData) error {
	if len(parts) != 3 {
		return errors.Errorf("time calibration should be time:<rate_id>:<sample_interval_fs>, got only %d colon-separated parts", len(parts))
	}
	rate, err := strconv.ParseUint(parts[1], 16, 8)
	if err != nil {
		return errors.Errorf("invalid sample rate ID %q, need a hexadecimal byte", parts[1])
	}
	if got := sampleIDToRate[rateID(rate)]; got == 0 {
		var rates []string
		for k := range sampleIDToRate {
			rates = append(rates, strconv.FormatInt(int64(k), 16))
		}
		return errors.Errorf("invalid sample rate ID %q, want one of %v", parts[1], rates)
	}
	dur, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return errors.Errorf("invalid duration %q, want a decimal number of femtoseconds", dur)
	}
	cal.time[rateID(rate)] = scope.Duration(dur)
	return nil
}

// readCalibrationLine reads a line from the config file and sets an appropriate value in calibration data.
func readCalibrationLine(s string, cal *calData) error {
	if len(s) == 0 || s[0] == '#' {
		return nil
	}
	parts := strings.Split(s, ":")
	switch parts[0] {
	case "volt":
		return parseVoltageCal(parts, cal)
	case "time":
		return parseTimeCal(parts, cal)
	}
	return errors.Errorf("Invalid line id %q, need volt or time", parts[0])
}

// readCalibrationFile reads the config file into the calData struct. cal maps must be non-nil.
func readCalibrationFile(cal *calData) error {
	cfgDir := configdir.New("goscope", "goscope")
	cfg := cfgDir.QueryFolderContainsFile(calDataFile)
	if cfg == nil {
		return errors.Errorf("calibration data file %q does not exist", calDataFile)
	}
	f, err := cfg.Open(calDataFile)
	if err != nil {
		return errors.Errorf("failed to open calibration data file %q: %v", calDataFile, err)
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	num := 0
	for s.Scan() {
		if err := readCalibrationLine(s.Text(), cal); err != nil {
			log.Printf("calibration data: syntax error in line %d: %v", num+1, err)
		}
		num++
	}
	if err := s.Err(); err != nil {
		return errors.Errorf("failed to read calibration data file %q: %v", calDataFile, err)
	}
	return nil
}

// getCalibration returns calibration data if found, or defaults if no data is present.
func getCalibration() *calData {
	cal := &calData{
		voltage: [2]map[rangeID]calVoltage{make(map[rangeID]calVoltage), make(map[rangeID]calVoltage)},
		time:    make(map[rateID]scope.Duration),
	}
	for id := range rangeIDToVolts {
		cal.voltage[0][id] = calVoltage{id.volts() * -1, 2 * id.volts() / 256}
		cal.voltage[1][id] = calVoltage{id.volts() * -1, 2 * id.volts() / 256}
	}
	for id, rate := range sampleIDToRate {
		cal.time[id] = rate.Interval()
	}
	if err := readCalibrationFile(cal); err != nil {
		log.Printf("No calibration data: %v, using defaults.", err)
	}
	return cal
}
