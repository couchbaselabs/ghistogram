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

// Package ghistogram provides a simple histogram of ints.  Unlike
// other histogram implementations, ghistogram avoids heap allocations
// (garbage creation) during data processing.
package ghistogram

import (
	"math"
)

// GHistogram is a simple int histogram implementation.  Unlike other
// histogram implementations, ghistogram avoids heap allocations
// (garbage creation) during data processing.
type GHistogram struct {
	// Bins are split across ranges and counts, where len(ranges) ==
	// len(counts).

	ranges []int // Lower bound of bin, so ranges[0] == binStart.
	counts []uint64
}

// NewGHistogram creates a new GHistogram.  The numBins must be >= 2.
// The binFirst is the width of the first bin.  The binGrowthFactor
// must be > 1.0.
func NewGHistogram(
	numBins int,
	binFirst int,
	binGrowthFactor float64) *GHistogram {
	gh := &GHistogram{
		ranges: make([]int, numBins),
		counts: make([]uint64, numBins),
	}

	gh.ranges[0] = 0
	gh.ranges[1] = binFirst

	for i := 2; i < len(gh.ranges); i++ {
		gh.ranges[i] =
			int(math.Ceil(binGrowthFactor * float64(gh.ranges[i-1])))
	}

	return gh
}

// Add increases the count in the bin for the given dataPoint.
func (gh *GHistogram) Add(dataPoint int, count uint64) {
	idx := search(gh.ranges, dataPoint)
	if idx >= 0 {
		gh.counts[idx] += count
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
