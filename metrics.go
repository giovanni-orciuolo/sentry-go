package sentry

import (
	"fmt"
	"hash/crc32"
	"math"
	"regexp"
	"sort"
	"strings"
)

type (
	NumberOrString interface {
		int | string
	}

	void struct{}
)

var (
	member     void
	keyRegex   = regexp.MustCompile(`[^a-zA-Z0-9_/.-]+`)
	valueRegex = regexp.MustCompile(`[^\w\d\s_:/@\.{}\[\]$-]+`)
	unitRegex  = regexp.MustCompile(`[^a-z]+`)
)

type MetricUnit struct {
	unit string
}

func (m MetricUnit) toString() string {
	return m.unit
}

func NanoSecond() MetricUnit {
	return MetricUnit{
		"nanosecond",
	}
}

func MicroSecond() MetricUnit {
	return MetricUnit{
		"microsecond",
	}
}

func MilliSecond() MetricUnit {
	return MetricUnit{
		"millisecond",
	}
}

func Second() MetricUnit {
	return MetricUnit{
		"second",
	}
}

func Minute() MetricUnit {
	return MetricUnit{
		"minute",
	}
}

func Hour() MetricUnit {
	return MetricUnit{
		"hour",
	}
}

func Day() MetricUnit {
	return MetricUnit{
		"day",
	}
}

func Week() MetricUnit {
	return MetricUnit{
		"week",
	}
}

func Bit() MetricUnit {
	return MetricUnit{
		"bit",
	}
}

func Byte() MetricUnit {
	return MetricUnit{
		"byte",
	}
}

func KiloByte() MetricUnit {
	return MetricUnit{
		"kilobyte",
	}
}

func KibiByte() MetricUnit {
	return MetricUnit{
		"kibibyte",
	}
}

func MegaByte() MetricUnit {
	return MetricUnit{
		"megabyte",
	}
}

func MebiByte() MetricUnit {
	return MetricUnit{
		"mebibyte",
	}
}

func GigaByte() MetricUnit {
	return MetricUnit{
		"gigabyte",
	}
}

func GibiByte() MetricUnit {
	return MetricUnit{
		"gibibyte",
	}
}

func TeraByte() MetricUnit {
	return MetricUnit{
		"terabyte",
	}
}

func TebiByte() MetricUnit {
	return MetricUnit{
		"tebibyte",
	}
}

func PetaByte() MetricUnit {
	return MetricUnit{
		"petabyte",
	}
}

func PebiByte() MetricUnit {
	return MetricUnit{
		"pebibyte",
	}
}

func ExaByte() MetricUnit {
	return MetricUnit{
		"exabyte",
	}
}

func ExbiByte() MetricUnit {
	return MetricUnit{
		"exbibyte",
	}
}

func Ratio() MetricUnit {
	return MetricUnit{
		"ratio",
	}
}

func Percent() MetricUnit {
	return MetricUnit{
		"percent",
	}
}

func CustomUnit(unit string) MetricUnit {
	return MetricUnit{
		unitRegex.ReplaceAllString(unit, ""),
	}
}

type Metric interface {
	GetType() string
	GetTags() map[string]string
	GetKey() string
	GetUnit() string
	GetTimestamp() int64
	GetWeight() int
	SerializeValue() string
	SerializeTags() string
	Add(value interface{})
}

type abstractMetric struct {
	key  string
	unit MetricUnit
	tags map[string]string
	// A unix timestamp (full seconds elapsed since 1970-01-01 00:00 UTC).
	timestamp int64
}

func (am *abstractMetric) GetTags() map[string]string {
	return am.tags
}

func (am *abstractMetric) GetKey() string {
	return am.key
}

func (am *abstractMetric) GetUnit() string {
	return am.unit.toString()
}

func (am *abstractMetric) GetTimestamp() int64 {
	return am.timestamp
}

func (am *abstractMetric) SerializeTags() string {
	return serializeTags(am.tags)
}

// Counter Metric.
type CounterMetric struct {
	value float64
	abstractMetric
}

func (c *CounterMetric) Add(value interface{}) {
	v := value.(float64)
	c.value += v
}

func (c *CounterMetric) GetType() string {
	return "c"
}

func (c *CounterMetric) GetWeight() int {
	return 1
}

func (c *CounterMetric) SerializeValue() string {
	return fmt.Sprintf(":%v", c.value)
}

// timestamp: A unix timestamp (full seconds elapsed since 1970-01-01 00:00 UTC).
func NewCounterMetric(key string, unit MetricUnit, tags map[string]string, timestamp int64, value float64) *CounterMetric {
	am := abstractMetric{
		key,
		unit,
		tags,
		timestamp,
	}

	return &CounterMetric{
		value,
		am,
	}
}

// Distribution Metric.
type DistributionMetric struct {
	values []float64
	abstractMetric
}

func (d *DistributionMetric) Add(value interface{}) {
	v := value.(float64)
	d.values = append(d.values, v)
}

func (d *DistributionMetric) GetType() string {
	return "d"
}

func (d *DistributionMetric) GetWeight() int {
	return len(d.values)
}

func (d *DistributionMetric) SerializeValue() string {
	var sb strings.Builder
	for _, el := range d.values {
		sb.WriteString(fmt.Sprintf(":%v", el))
	}
	return sb.String()
}

// timestamp: A unix timestamp (full seconds elapsed since 1970-01-01 00:00 UTC).
func NewDistributionMetric(key string, unit MetricUnit, tags map[string]string, timestamp int64, value float64) *DistributionMetric {
	am := abstractMetric{
		key,
		unit,
		tags,
		timestamp,
	}

	return &DistributionMetric{
		[]float64{value},
		am,
	}
}

// Gauge Metric.
type GaugeMetric struct {
	last  float64
	min   float64
	max   float64
	sum   float64
	count float64
	abstractMetric
}

func (g *GaugeMetric) Add(value interface{}) {
	v := value.(float64)
	g.last = v
	g.min = math.Min(g.min, v)
	g.max = math.Max(g.max, v)
	g.sum += v
	g.count++
}

func (g *GaugeMetric) GetType() string {
	return "g"
}

func (g *GaugeMetric) GetWeight() int {
	return 5
}

func (g *GaugeMetric) SerializeValue() string {
	return fmt.Sprintf(":%v:%v:%v:%v:%v", g.last, g.min, g.max, g.sum, g.count)
}

// timestamp: A unix timestamp (full seconds elapsed since 1970-01-01 00:00 UTC).
func NewGaugeMetric(key string, unit MetricUnit, tags map[string]string, timestamp int64, value float64) *GaugeMetric {
	am := abstractMetric{
		key,
		unit,
		tags,
		timestamp,
	}

	return &GaugeMetric{
		value, // last
		value, // min
		value, // max
		value, // sum
		value, // count
		am,
	}
}

// Set Metric.
type SetMetric struct {
	values map[int]void
	abstractMetric
}

func (s *SetMetric) Add(value interface{}) {
	v := value.(int)
	s.values[v] = member
}

func (s *SetMetric) GetType() string {
	return "s"
}

func (s *SetMetric) GetWeight() int {
	return len(s.values)
}

func (s *SetMetric) SerializeValue() string {
	values := make([]int, 0, len(s.values))
	for k := range s.values {
		values = append(values, k)
	}
	sortSlice(values)

	var sb strings.Builder
	for _, el := range values {
		sb.WriteString(fmt.Sprintf(":%v", el))
	}

	return sb.String()
}

func setStringKeyToInt(s string) uint32 {
	return crc32.ChecksumIEEE([]byte(s))
}

// timestamp: A unix timestamp (full seconds elapsed since 1970-01-01 00:00 UTC).
func NewSetMetric[T NumberOrString](key string, unit MetricUnit, tags map[string]string, timestamp int64, value int) *SetMetric {
	am := abstractMetric{
		key,
		unit,
		tags,
		timestamp,
	}

	return &SetMetric{
		map[int]void{
			value: member,
		},
		am,
	}
}

func sanitizeKey(s string) string {
	return keyRegex.ReplaceAllString(s, "_")
}

func sanitizeValue(s string) string {
	return valueRegex.ReplaceAllString(s, "")
}

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
}

func sortSlice[T Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

func serializeTags(tags map[string]string) string {
	var sb strings.Builder

	values := make([]string, 0, len(tags))
	for k := range tags {
		values = append(values, k)
	}
	sortSlice(values)

	for _, key := range values {
		val := sanitizeValue(tags[key])
		key = sanitizeKey(key)
		sb.WriteString(fmt.Sprintf("%s:%s,", key, val))
	}
	s := sb.String()
	if len(s) > 0 {
		s = s[:len(s)-1]
	}
	return s
}
