//  Copyright (c) 2015 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

// Package ghistogram provides a simple histogram of ints that avoids
// heap allocations (garbage creation) during data processing.
package ghistogram

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
)

// Histogram is a simple int histogram implementation that avoids heap
// allocations (garbage creation) during its processing of incoming
// data points.
//
// The histogram bins are split across arrays of Ranges and Counts,
// where len(Ranges) == len(Counts).  These arrays are public in case
// users wish to use reflection or JSON marhsallings.
//
// An optional growth factor for bin sizes is supported - see
// NewHistogram() binGrowthFactor parameter.
//
// Concurrent access (e.g., locking) on a Histogram is a
// responsibility of the user's application.
type Histogram struct {
	Ranges []uint64 // Lower bound of bin, so Ranges[0] == binStart.
	Counts []uint64

	MinDataPoint uint64
	MaxDataPoint uint64
}

// NewHistogram creates a new, ready to use Histogram.  The numBins
// must be >= 2.  The binFirst is the width of the first bin.  The
// binGrowthFactor must be > 1.0 or 0.0.
//
// A special case of binGrowthFactor of 0.0 means the the allocated
// bins will have constant, non-growing size or "width".
func NewHistogram(
	numBins int,
	binFirst uint64,
	binGrowthFactor float64) *Histogram {
	gh := &Histogram{
		Ranges: make([]uint64, numBins),
		Counts: make([]uint64, numBins),

		MinDataPoint: math.MaxUint64,
		MaxDataPoint: 0,
	}

	gh.Ranges[0] = 0
	gh.Ranges[1] = binFirst

	for i := 2; i < len(gh.Ranges); i++ {
		if binGrowthFactor == 0.0 {
			gh.Ranges[i] = gh.Ranges[i-1] + binFirst
		} else {
			gh.Ranges[i] =
				uint64(math.Ceil(binGrowthFactor * float64(gh.Ranges[i-1])))
		}
	}

	return gh
}

// Add increases the count in the bin for the given dataPoint.
func (gh *Histogram) Add(dataPoint uint64, count uint64) {
	idx := search(gh.Ranges, dataPoint)
	if idx >= 0 {
		gh.Counts[idx] += count
	}
	if gh.MinDataPoint > count {
		gh.MinDataPoint = count
	}
	if gh.MaxDataPoint < count {
		gh.MaxDataPoint = count
	}
}

// Finds the last arr index where the arr entry <= dataPoint.
func search(arr []uint64, dataPoint uint64) int {
	i, j := 0, len(arr)

	for i < j {
		h := i + (j-i)/2 // Avoids h overflow, where i <= h < j.
		if dataPoint >= arr[h] {
			i = h + 1
		} else {
			j = h
		}
	}

	return i - 1
}

// AddAll adds all the Counts from the src histogram into this
// histogram.  The src and this histogram must either have the same
// exact creation parameters.
func (gh *Histogram) AddAll(src *Histogram) {
	for i := 0; i < len(src.Counts); i++ {
		gh.Counts[i] += src.Counts[i]
	}
	if gh.MinDataPoint > src.MinDataPoint {
		gh.MinDataPoint = src.MinDataPoint
	}
	if gh.MaxDataPoint < src.MaxDataPoint {
		gh.MaxDataPoint = src.MaxDataPoint
	}
}

// Graph emits an ascii graph to the optional bufOut, allocating a
// bufOut if none is supplied.  Returns the bufOut.  Each line emitted
// will have the given, optional prefix.
//
// For example:
//       0+  10=2 10.00% ********
//      10+  10=1 10.00% ****
//      20+  10=3 10.00% ************
func (gh *Histogram) EmitGraph(prefix []byte,
	bufOut *bytes.Buffer) *bytes.Buffer {
	ranges := gh.Ranges
	rangesN := len(ranges)
	counts := gh.Counts
	countsN := len(counts)

	if bufOut == nil {
		bufOut = bytes.NewBuffer(make([]byte, 0, 80*countsN))
	}

	var totCount uint64
	var maxCount uint64
	for _, c := range counts {
		totCount += c
		if maxCount < c {
			maxCount = c
		}
	}
	totCountF := float64(totCount)
	maxCountF := float64(maxCount)

	widthRange := len(strconv.Itoa(int(ranges[rangesN-1])))
	widthWidth := len(strconv.Itoa(int(ranges[rangesN-1] - ranges[rangesN-2])))
	widthCount := len(strconv.Itoa(int(maxCount)))

	// Each line looks like: "[prefix]START+WIDTH=COUNT PCT% BAR\n"
	f := fmt.Sprintf("%%%dd+%%%dd=%%%dd%% 7.2f%%%%",
		widthRange, widthWidth, widthCount)

	var runCount uint64 // Running total while emitting lines.

	barLen := float64(len(bar))

	for i, c := range counts {
		if prefix != nil {
			bufOut.Write(prefix)
		}

		var width uint64
		if i < countsN-1 {
			width = uint64(ranges[i+1] - ranges[i])
		}

		runCount += c
		fmt.Fprintf(bufOut, f, ranges[i], width, c,
			100.0*(float64(runCount)/totCountF))

		if c > 0 {
			bufOut.Write([]byte(" "))
			barWant := int(math.Floor(barLen * (float64(c) / maxCountF)))
			bufOut.Write(bar[0:barWant])
		}

		bufOut.Write([]byte("\n"))
	}

	return bufOut
}

var bar = []byte("******************************")
