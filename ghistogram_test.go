package ghistogram

import (
	"testing"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		arr []int
		val int
		exp int
	}{
		{[]int(nil), 0, -1},
		{[]int(nil), 100, -1},
		{[]int(nil), -100, -1},

		{[]int{0}, 0, 0},
		{[]int{0, 10}, 0, 0},
		{[]int{0, 10, 20}, 0, 0},

		{[]int{0}, 1, 0},
		{[]int{0, 10}, 1, 0},
		{[]int{0, 10, 20}, 1, 0},

		{[]int{0}, 10, 0},
		{[]int{0, 10}, 10, 1},
		{[]int{0, 10, 20}, 10, 1},

		{[]int{0}, 15, 0},
		{[]int{0, 10}, 15, 1},
		{[]int{0, 10, 20}, 15, 1},

		{[]int{0}, 20, 0},
		{[]int{0, 10}, 20, 1},
		{[]int{0, 10, 20}, 20, 2},

		{[]int{0}, 30, 0},
		{[]int{0, 10}, 30, 1},
		{[]int{0, 10, 20}, 30, 2},
	}

	for testi, test := range tests {
		got := search(test.arr, test.val)
		if got != test.exp {
			t.Errorf("test #%d, arr: %v, val: %d, exp: %d, got: %d",
				testi, test.arr, test.val, test.exp, got)
		}
		if got >= 0 &&
			got < len(test.arr) &&
			test.arr[got] > test.val {
			t.Errorf("test #%d, test.arr[got] > test.val,"+
				" arr: %v, val: %d, exp: %d, got: %d",
				testi, test.arr, test.val, test.exp, got)
		}
	}
}

func TestNewHistogram(t *testing.T) {
	tests := []struct {
		numBins         int
		binFirst        int
		binGrowthFactor float64
		exp             []int
	}{
		{2, 123, 10.0, []int{0, 123}},
		{2, 123, 10.0, []int{0, 123}},

		{5, 10, 2.0, []int{0, 10, 20, 40, 80}},
		{5, 10, 1.5, []int{0, 10, 15, 23, 35}},
	}

	for testi, test := range tests {
		gh := NewHistogram(
			test.numBins, test.binFirst, test.binGrowthFactor)
		if len(gh.Ranges) != len(gh.Counts) {
			t.Errorf("mismatched len's")
		}
		if len(gh.Ranges) != test.numBins {
			t.Errorf("wrong len's")
		}
		if len(gh.Ranges) != len(test.exp) {
			t.Errorf("unequal len's")
		}
		for i := 0; i < len(gh.Ranges); i++ {
			if gh.Ranges[i] != test.exp[i] {
				t.Errorf("test #%d, actual (%v) != exp (%v)",
					testi, gh.Ranges, test.exp)
			}
		}
	}
}

func TestAdd(t *testing.T) {
	// Bins will look like: {0, 10, 20, 40, 80}.
	gh := NewHistogram(5, 10, 2.0)

	tests := []struct {
		val int
		exp []uint64
	}{
		{0, []uint64{1, 0, 0, 0, 0}},
		{0, []uint64{2, 0, 0, 0, 0}},
		{0, []uint64{3, 0, 0, 0, 0}},

		{2, []uint64{4, 0, 0, 0, 0}},
		{3, []uint64{5, 0, 0, 0, 0}},
		{4, []uint64{6, 0, 0, 0, 0}},

		{10, []uint64{6, 1, 0, 0, 0}},
		{11, []uint64{6, 2, 0, 0, 0}},
		{12, []uint64{6, 3, 0, 0, 0}},

		{100, []uint64{6, 3, 0, 0, 1}},
		{90, []uint64{6, 3, 0, 0, 2}},
		{80, []uint64{6, 3, 0, 0, 3}},

		{20, []uint64{6, 3, 1, 0, 3}},
		{30, []uint64{6, 3, 2, 0, 3}},
		{40, []uint64{6, 3, 2, 1, 3}},
	}

	for testi, test := range tests {
		gh.Add(test.val, 1)

		for i := 0; i < len(gh.Counts); i++ {
			if gh.Counts[i] != test.exp[i] {
				t.Errorf("test #%d, actual (%v) != exp (%v)",
					testi, gh.Counts, test.exp)
			}
		}
	}
}

func TestAddAll(t *testing.T) {
	// Bins will look like: {0, 10, 20, 40, 80}.
	gh := NewHistogram(5, 10, 2.0)

	gh.Add(15, 2)
	gh.Add(25, 3)
	gh.Add(1000, 1)

	gh2 := NewHistogram(5, 10, 2.0)
	gh2.AddAll(gh)
	gh2.AddAll(gh)

	exp := []uint64{0, 4, 6, 0, 2}

	for i := 0; i < len(gh2.Counts); i++ {
		if gh2.Counts[i] != exp[i] {
			t.Errorf("AddAll mismatch, actual (%v) != exp (%v)",
				gh2.Counts, exp)
		}
	}
}
