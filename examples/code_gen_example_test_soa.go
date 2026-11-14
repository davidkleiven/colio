// Generated code by coliosoa. DO NOT EDIT

package examples

import (
	"github.com/davidkleiven/colio/pkg"
	"iter"
	"time"
)

type MeasurementSoa struct {
	Id    []string    `parquet:"id"`
	Time  []time.Time `parquet:"time"`
	Value []float64   `parqiet:"value"`
}

func (v *MeasurementSoa) Allocate(n int) {
	v.Id = make([]string, n)
	v.Time = make([]time.Time, n)
	v.Value = make([]float64, n)
}

func (v *MeasurementSoa) Columns() []pkg.Column {
	return []pkg.Column{
		{Name: "Id", Tags: `parquet:"id"`, Ptr: &v.Id},
		{Name: "Time", Tags: `parquet:"time"`, Ptr: &v.Time},
		{Name: "Value", Tags: `parqiet:"value"`, Ptr: &v.Value},
	}
}

func (v *MeasurementSoa) Cap() int {
	return min(
		len(v.Id),
		len(v.Time),
		len(v.Value),
	)
}

func (v *MeasurementSoa) ToAos() iter.Seq[Measurement] {
	return func(yield func(Measurement) bool) {
		for i := range v.Cap() {
			item := Measurement{
				Id:    v.Id[i],
				Time:  v.Time[i],
				Value: v.Value[i],
			}
			if !yield(item) {
				return
			}
		}
	}
}

func (v *MeasurementSoa) Delete(i int) {
	v.Id[i] = v.Id[len(v.Id)-1]
	v.Id = v.Id[:len(v.Id)-1]
	v.Time[i] = v.Time[len(v.Time)-1]
	v.Time = v.Time[:len(v.Time)-1]
	v.Value[i] = v.Value[len(v.Value)-1]
	v.Value = v.Value[:len(v.Value)-1]
}

func (v *MeasurementSoa) FromAos(aos []Measurement) *MeasurementSoa {
	v.Allocate(len(aos))
	for i, item := range aos {
		v.Id[i] = item.Id
		v.Time[i] = item.Time
		v.Value[i] = item.Value
	}
	return v
}
