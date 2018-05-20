// +build go1.7

package statsd

import (
	"fmt"
	"strconv"
	"testing"
)

var statBytes []byte
var stat string

// Results:
// BenchmarkStatBuildGauge_Sprintf-8       	     500	  45699958 ns/op
// BenchmarkStatBuildGauge_Concat-8        	    1000	  23452863 ns/op
// BenchmarkStatBuildGauge_BytesAppend-8   	    1000	  21705121 ns/op
func BenchmarkStatBuildGauge_Sprintf(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for x := 0; x < 100000; x++ {
			stat = fmt.Sprintf("%f|g", 3.14159)
		}
	}
}

func BenchmarkStatBuildGauge_Concat(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for x := 0; x < 100000; x++ {
			stat = strconv.FormatFloat(3.14159, 'f', -1, 64) + "|g"
		}
	}
}

func BenchmarkStatBuildGauge_BytesAppend(b *testing.B) {
	suffix := []byte("|g")

	for n := 0; n < b.N; n++ {
		for x := 0; x < 100000; x++ {
			statBytes = []byte{}
			statBytes = append(strconv.AppendFloat(statBytes, 3.14159, 'f', -1, 64), suffix...)
		}
	}
}

func BenchmarkStatBuildCount_Sprintf(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for x := 0; x < 100000; x++ {
			stat = fmt.Sprintf("%d|c", 314)
		}
	}
}

func BenchmarkStatBuildCount_Concat(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for x := 0; x < 100000; x++ {
			stat = strconv.FormatInt(314, 10) + "|c"
		}
	}
}

func BenchmarkStatBuildCount_BytesAppend(b *testing.B) {
	suffix := []byte("|c")

	for n := 0; n < b.N; n++ {
		for x := 0; x < 100000; x++ {
			statBytes = []byte{}
			statBytes = append(strconv.AppendInt(statBytes, 314, 10), suffix...)
		}
	}
}

var FormatSink []byte

func BenchmarkClientFormat(b *testing.B) {
	var tests = []struct {
		globalNamespace string
		globalTags      []string
		name            string
		value           interface{}
		suffix          string
		tags            []string
	}{
		{"", nil, "test.gauge", 1.0, gaugeSuffix, nil},
		{"", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA"}},
		{"", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
		{"", nil, "test.count", int64(1), countSuffix, []string{"tagA"}},
		{"", nil, "test.histogram", 2.3, histogramSuffix, []string{"tagA"}},
		{"", nil, "test.distribution", 2.3, distributionSuffix, []string{"tagA"}},
		{"", nil, "test.set", "uuid", setSuffix, []string{"tagA"}},
		{"flubber.", nil, "test.set", "uuid", setSuffix, []string{"tagA"}},
		{"", []string{"tagC"}, "test.set", "uuid", setSuffix, []string{"tagA"}},
		{"", nil, "test.count", int64(1), countSuffix, []string{"hello\nworld"}},
	}

	b.ReportAllocs()

	for i, tt := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			c := &Client{
				Namespace: tt.globalNamespace,
				Tags:      tt.globalTags,
			}

			b.ReportAllocs()
			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				c.appendStat(tt.name, tt.value, tt.suffix, 1.0, tt.tags...)
			}
		})
	}
}

func BenchmarkFlush(b *testing.B) {
	var tests = []struct {
		globalNamespace string
		globalTags      []string
		name            string
		value           interface{}
		suffix          string
		tags            []string
	}{
		{"", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
	}

	b.ReportAllocs()

	for i, tt := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			c, err := New(ConnAddr("127.0.0.1:56789"))
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					c.send(tt.name, tt.value, tt.suffix, 1.0, tt.tags...)
				}
			})
		})
	}
}

func BenchmarkFlushBatch(b *testing.B) {
	var tests = []struct {
		batchSize       int
		globalNamespace string
		globalTags      []string
		name            string
		value           interface{}
		suffix          string
		tags            []string
	}{
		{1, "", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
		{5, "", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
		{10, "", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
		{100, "", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
		{1000, "", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
	}

	b.ReportAllocs()

	for _, tt := range tests {
		b.Run(strconv.Itoa(tt.batchSize), func(b *testing.B) {
			c, err := New(ConnAddr("127.0.0.1:56789"), ConnBuffer(tt.batchSize))
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					c.send(tt.name, tt.value, tt.suffix, 1.0, tt.tags...)
				}
			})
		})
	}
}

func BenchmarkFlushBatchShard(b *testing.B) {
	var tests = []struct {
		batchSize       int
		globalNamespace string
		globalTags      []string
		name            string
		value           interface{}
		suffix          string
		tags            []string
	}{
		{1, "", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
		{5, "", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
		{10, "", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
		{100, "", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
		{1000, "", nil, "test.gauge", 1.0, gaugeSuffix, []string{"tagA", "tagB"}},
	}

	b.ReportAllocs()

	for _, tt := range tests {
		b.Run(strconv.Itoa(tt.batchSize), func(b *testing.B) {
			c, err := New(ConnAddr("127.0.0.1:56789"), ConnBuffer(tt.batchSize))
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				c, err := c.Clone()
				if err != nil {
					b.Fatal(err)
				}
				for pb.Next() {
					c.send(tt.name, tt.value, tt.suffix, 1.0, tt.tags...)
				}
			})

			b.StopTimer()
			c.Close()
		})
	}
}