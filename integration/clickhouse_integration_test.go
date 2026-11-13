package integration

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
	hch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/davidkleiven/colio/clickhouse"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tch "github.com/testcontainers/testcontainers-go/modules/clickhouse"
)

func setupChContainer(t testing.TB) ch.Options {
	t.Helper()
	ctx := context.Background()

	ctr, err := tch.Run(ctx,
		"clickhouse/clickhouse-server:26.10.3.7-alpine",
		tch.WithUsername("default"),
		tch.WithPassword("clickhouse"),
		tch.WithDatabase("testdb"),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, ctr) // auto-terminate on test end

	// Get connection info
	host, err := ctr.ConnectionHost(ctx)
	require.NoError(t, err)
	return ch.Options{
		Address:  host,
		Database: "testdb",
		User:     "default",
		Password: "clickhouse",
	}
}

const createPersonsSql = `
	CREATE TABLE persons (
    name        String,
    age         Int64,
    num_children Int32,
    height      Float64,
    weight      Float32,
    created_at   DateTime64(6)
) ENGINE = MergeTree()
ORDER BY created_at;
`

func TestReadClickhouseIntegrationTest(t *testing.T) {
	opts := setupChContainer(t)
	ctx := context.Background()
	c, err := ch.Dial(ctx, opts)
	require.NoError(t, err)
	err = c.Do(ctx, ch.Query{Body: createPersonsSql})

	rows := []clickhouse.Person{
		{Name: "Alice", Age: 34, NumChildren: 2, Height: 1.68, Weight: 62.5, CreatedAt: time.Now().UTC()},
		{Name: "Bob", Age: 28, NumChildren: 0, Height: 1.82, Weight: 78.0, CreatedAt: time.Now().UTC()},
		{Name: "Charlie", Age: 45, NumChildren: 3, Height: 1.75, Weight: 85.2, CreatedAt: time.Now().Add(-24 * time.Hour).UTC()},
		{Name: "Diana", Age: 31, NumChildren: 1, Height: 1.60, Weight: 54.3, CreatedAt: time.Now().Add(-48 * time.Hour).UTC()},
		{Name: "Eve", Age: 52, NumChildren: 4, Height: 1.70, Weight: 70.1, CreatedAt: time.Now().Add(-72 * time.Hour).UTC()},
	}
	var (
		names       proto.ColStr
		ages        proto.ColInt64
		numChildren proto.ColInt32
		heights     proto.ColFloat64
		weights     proto.ColFloat32
		createdAts  proto.ColDateTime64
	)
	createdAts.WithPrecision(proto.PrecisionMicro)

	for i := range rows {
		names.Append(rows[i].Name)
		ages.Append(int64(rows[i].Age))
		numChildren.Append(rows[i].NumChildren)
		heights.Append(rows[i].Height)
		weights.Append(rows[i].Weight)
		createdAts.Append(rows[i].CreatedAt)
	}
	require.NoError(t, err)

	err = c.Do(ctx, ch.Query{
		Body: "INSERT INTO persons (name, age, num_children, height, weight, created_at) VALUES",
		Input: proto.Input{
			{Name: "name", Data: &names},
			{Name: "age", Data: &ages},
			{Name: "num_children", Data: &numChildren},
			{Name: "height", Data: &heights},
			{Name: "weight", Data: &weights},
			{Name: "created_at", Data: &createdAts},
		},
	})

	require.NoError(t, err)

	var result clickhouse.PersonSoa
	err = clickhouse.NewReader(c, "SELECT * FROM persons").Bind(&result).Read[*clickhouse.PersonSoa](ctx, nil)
	require.ErrorIs(t, err, io.EOF)
	require.Equal(t, 5, result.Cap())
}

func BenchmarkColVsHighlevelApi(b *testing.B) {
	opts := setupChContainer(b)
	ctx := context.Background()
	client, err := ch.Dial(ctx, opts)
	require.NoError(b, err)
	err = client.Do(ctx, ch.Query{Body: createPersonsSql})
	require.NoError(b, err)

	num := 2_000_000
	data := make([]clickhouse.Person, 0, num)
	for i := range num {
		person := clickhouse.Person{
			CreatedAt:   time.Now(),
			NumChildren: int32(i % 4),
			Height:      float64(i % 200),
			Weight:      float32(i % 80),
		}
		data = append(data, person)
	}

	hOpts := hch.Options{
		Addr: []string{opts.Address},
		Auth: hch.Auth{Username: opts.User, Password: opts.Password},
	}
	conn, err := hch.Open(&hOpts)
	require.NoError(b, err)

	batch, err := conn.PrepareBatch(ctx, "INSERT INTO testdb.persons")
	require.NoError(b, err)

	for _, v := range data {
		err := batch.AppendStruct(&v)
		require.NoError(b, err)
	}

	err = batch.Send()
	require.NoError(b, err)

	b.Run("high-level-api", func(b *testing.B) {
		for range b.N {
			var result []clickhouse.Person
			err := conn.Select(ctx, &result, "SELECT * FROM testdb.persons")
			require.NoError(b, err)
		}
	})

	b.Run("col-based-api", func(b *testing.B) {
		for range b.N {
			var result clickhouse.PersonSoa
			err := clickhouse.NewReader(client, "SELECT * FROM testdb.persons").Bind(&result).Read[*clickhouse.PersonSoa](ctx, nil)
			require.ErrorIs(b, err, io.EOF)
		}
	})
}
