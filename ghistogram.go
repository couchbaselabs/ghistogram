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
	"math"
)

// Histogram is a simple int histogram implementation that avoids heap
// allocations (garbage creation) during its processing of incoming
// data points.
//
// The histogram bins are split across arrays of Ranges and Counts,
// where len(Ranges) == len(Counts).  These arrays are public in case
// users wish to use reflection or JSON marhsallings.
//
// Concurrent access (e.g., locking) on a Histogram is a
// responsibility of the user's application.
type Histogram struct {
	Ranges []int // Lower bound of bin, so Ranges[0] == binStart.
	Counts []uint64
}

// NewHistogram creates a new, ready to use Histogram.  The numBins
// must be >= 2.  The binFirst is the width of the first bin.  The
// binGrowthFactor must be > 1.0.
func NewHistogram(
	numBins int,
	binFirst int,
	binGrowthFactor float64) *Histogram {
	gh := &Histogram{
		Ranges: make([]int, numBins),
		Counts: make([]uint64, numBins),
	}

	gh.Ranges[0] = 0
	gh.Ranges[1] = binFirst

	for i := 2; i < len(gh.Ranges); i++ {
		gh.Ranges[i] =
			int(math.Ceil(binGrowthFactor * float64(gh.Ranges[i-1])))
	}

	return gh
}

// Add increases the count in the bin for the given dataPoint.
func (gh *Histogram) Add(dataPoint int, count uint64) {
	idx := search(gh.Ranges, dataPoint)
	if idx >= 0 {
		gh.Counts[idx] += count
	}
}

// Finds the last arr index where the arr entry <= dataPoint.
func search(arr []int, dataPoint int) int {
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
}
