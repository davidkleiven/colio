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
