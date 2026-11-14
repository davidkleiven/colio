package clickhouse

import (
	"bytes"
	"slices"
	"testing"
	"time"

	"github.com/ClickHouse/ch-go/proto"
	"github.com/davidkleiven/colio/pkg"
	"github.com/stretchr/testify/require"
)

func TestDecodeBlockIntoSoa(t *testing.T) {
	createdAt := time.UnixMilli(1_700_000_000_123).UTC()

	name := new(proto.ColStr)
	name.Append("John")
	name.Append("Susan")

	age := new(proto.ColInt64)
	age.Append(52)
	age.Append(48)

	numChildren := new(proto.ColInt32)
	numChildren.Append(2)
	numChildren.Append(1)

	height := new(proto.ColFloat64)
	height.Append(180.0)
	height.Append(168.0)

	weight := new(proto.ColFloat32)
	weight.Append(82.0)
	weight.Append(68.0)

	created := new(proto.ColDateTime64).WithPrecision(proto.PrecisionMilli)
	created.Append(createdAt)
	created.Append(createdAt)

	// Deliberately not in struct order: columns are matched by name.
	block := proto.Block{Columns: 6, Rows: 2}
	var buf proto.Buffer
	err := block.EncodeBlock(&buf, proto.Version, []proto.InputColumn{
		{Name: "created_at", Data: created},
		{Name: "weight", Data: weight},
		{Name: "height", Data: height},
		{Name: "num_children", Data: numChildren},
		{Name: "age", Data: age},
		{Name: "name", Data: name},
	})
	require.NoError(t, err)

	var results proto.Results
	var decoded proto.Block
	reader := proto.NewReader(bytes.NewReader(buf.Buf))
	require.NoError(t, decoded.DecodeBlock(reader, proto.Version, results.Auto()))
	require.Equal(t, 2, results.Rows())

	var soa PersonSoa
	r := NewReader(nil, "SELECT * FROM person").Bind(&soa)
	require.NoError(t, r.Err)
	require.NoError(t, r.decode(results))

	local := createdAt.In(time.Local)
	require.Equal(t, []Person{
		{Name: "John", Age: 52, NumChildren: 2, Height: 180.0, Weight: 82.0, CreatedAt: local},
		{Name: "Susan", Age: 48, NumChildren: 1, Height: 168.0, Weight: 68.0, CreatedAt: local},
	}, slices.Collect(soa.ToAos()))
}

func TestDecodeListColumnsIntoSoa(t *testing.T) {
	moment := time.UnixMilli(1_700_000_000_123).UTC()

	name := new(proto.ColStr)
	name.Append("John")
	name.Append("Susan")

	tags := proto.NewArray[int64](new(proto.ColInt64))
	tags.Append([]int64{1, 2, 3})
	tags.Append(nil)

	names := proto.NewArray[string](new(proto.ColStr))
	names.Append([]string{"a", "b"})
	names.Append(nil)

	steps := proto.NewArray[float64](new(proto.ColFloat64))
	steps.Append([]float64{1.5, 2.5})
	steps.Append(nil)

	times := proto.NewArray[time.Time](new(proto.ColDateTime64).WithPrecision(proto.PrecisionMilli))
	times.Append([]time.Time{moment})
	times.Append(nil)

	// Deliberately not in struct order: columns are matched by name.
	block := proto.Block{Columns: 5, Rows: 2}
	var buf proto.Buffer
	err := block.EncodeBlock(&buf, proto.Version, []proto.InputColumn{
		{Name: "times", Data: times},
		{Name: "steps", Data: steps},
		{Name: "names", Data: names},
		{Name: "tags", Data: tags},
		{Name: "name", Data: name},
	})
	require.NoError(t, err)

	var results proto.Results
	var decoded proto.Block
	reader := proto.NewReader(bytes.NewReader(buf.Buf))
	require.NoError(t, decoded.DecodeBlock(reader, proto.Version, results.Auto()))
	require.Equal(t, 2, results.Rows())

	var soa EventSoa
	r := NewReader(nil, "SELECT * FROM event").Bind(&soa)
	require.NoError(t, r.Err)
	require.NoError(t, r.decode(results))

	// Empty arrays come back as empty lists, matching what the parquet decoder
	// produces for empty lists. Times are returned in the local zone.
	local := moment.In(time.Local)
	require.Equal(t, []Event{
		{Name: "John", Tags: []int{1, 2, 3}, Names: []string{"a", "b"}, Steps: []float64{1.5, 2.5}, Times: []time.Time{local}},
		{Name: "Susan", Tags: []int{}, Names: []string{}, Steps: []float64{}, Times: []time.Time{}},
	}, slices.Collect(soa.ToAos()))
}

func TestDecodeListFromWrongColumnType(t *testing.T) {
	store, err := newStore(new([][]int64))
	require.NoError(t, err)
	require.ErrorContains(t, store(new(proto.ColInt64)), "cannot decode Int64 into *[][]int64")
}

func TestBindUnsupportedType(t *testing.T) {
	var dst unsupported
	r := NewReader(nil, "").Bind(&dst)
	require.Error(t, r.Err)
}

type unsupported struct {
	Values []uint64
}

func (u *unsupported) Columns() []pkg.Column {
	return []pkg.Column{{Name: "Values", Ptr: &u.Values}}
}

func (u *unsupported) Cap() int { return len(u.Values) }

func TestDecodeWrongColumnType(t *testing.T) {
	age := new(proto.ColInt64)
	age.Append(52)

	var results proto.Results
	results = append(results, proto.ResultColumn{Name: "name", Data: age})

	var soa PersonSoa
	r := NewReader(nil, "").Bind(&soa)
	require.NoError(t, r.Err)
	require.ErrorContains(t, r.decode(results), "cannot decode")
}
