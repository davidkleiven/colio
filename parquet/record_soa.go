// Generated code by coliosoa. DO NOT EDIT

package parquet

import (
	"github.com/davidkleiven/colio/colio"
	"iter"
	"time"
)

type PersonSoa struct {
	Name        []string    `parquet:"name"`
	Age         []int       `parquet:"age"`
	NumChildren []int32     `parquet:"numChildren"`
	Height      []float64   `parquet:"height"`
	Weight      []float32   `parquet:"weight"`
	CreatedAt   []time.Time `parquet:"createdAt"`
}

func (v *PersonSoa) Allocate(n int) {
	v.Name = make([]string, n)
	v.Age = make([]int, n)
	v.NumChildren = make([]int32, n)
	v.Height = make([]float64, n)
	v.Weight = make([]float32, n)
	v.CreatedAt = make([]time.Time, n)
}

func (v *PersonSoa) Columns() []colio.Column {
	return []colio.Column{
		{Name: "Name", Tags: `parquet:"name"`, Ptr: &v.Name},
		{Name: "Age", Tags: `parquet:"age"`, Ptr: &v.Age},
		{Name: "NumChildren", Tags: `parquet:"numChildren"`, Ptr: &v.NumChildren},
		{Name: "Height", Tags: `parquet:"height"`, Ptr: &v.Height},
		{Name: "Weight", Tags: `parquet:"weight"`, Ptr: &v.Weight},
		{Name: "CreatedAt", Tags: `parquet:"createdAt"`, Ptr: &v.CreatedAt},
	}
}

func (v *PersonSoa) Cap() int {
	return min(
		len(v.Name),
		len(v.Age),
		len(v.NumChildren),
		len(v.Height),
		len(v.Weight),
		len(v.CreatedAt),
	)
}

func (v *PersonSoa) ToAos() iter.Seq[Person] {
	return func(yield func(Person) bool) {
		for i := range v.Cap() {
			item := Person{
				Name:        v.Name[i],
				Age:         v.Age[i],
				NumChildren: v.NumChildren[i],
				Height:      v.Height[i],
				Weight:      v.Weight[i],
				CreatedAt:   v.CreatedAt[i],
			}
			if !yield(item) {
				return
			}
		}
	}
}

func (v *PersonSoa) Delete(i int) {
	v.Name[i] = v.Name[len(v.Name)-1]
	v.Name = v.Name[:len(v.Name)-1]
	v.Age[i] = v.Age[len(v.Age)-1]
	v.Age = v.Age[:len(v.Age)-1]
	v.NumChildren[i] = v.NumChildren[len(v.NumChildren)-1]
	v.NumChildren = v.NumChildren[:len(v.NumChildren)-1]
	v.Height[i] = v.Height[len(v.Height)-1]
	v.Height = v.Height[:len(v.Height)-1]
	v.Weight[i] = v.Weight[len(v.Weight)-1]
	v.Weight = v.Weight[:len(v.Weight)-1]
	v.CreatedAt[i] = v.CreatedAt[len(v.CreatedAt)-1]
	v.CreatedAt = v.CreatedAt[:len(v.CreatedAt)-1]
}

func (v *PersonSoa) FromAos(aos []Person) *PersonSoa {
	v.Allocate(len(aos))
	for i, item := range aos {
		v.Name[i] = item.Name
		v.Age[i] = item.Age
		v.NumChildren[i] = item.NumChildren
		v.Height[i] = item.Height
		v.Weight[i] = item.Weight
		v.CreatedAt[i] = item.CreatedAt
	}
	return v
}

type EventSoa struct {
	Name  []string      `parquet:"name"`
	Tags  [][]int       `parquet:"tags,list"`
	Steps [][]float64   `parquet:"steps"`
	Times [][]time.Time `parquet:"times,list"`
}

func (v *EventSoa) Allocate(n int) {
	v.Name = make([]string, n)
	v.Tags = make([][]int, n)
	v.Steps = make([][]float64, n)
	v.Times = make([][]time.Time, n)
}

func (v *EventSoa) Columns() []colio.Column {
	return []colio.Column{
		{Name: "Name", Tags: `parquet:"name"`, Ptr: &v.Name},
		{Name: "Tags", Tags: `parquet:"tags,list"`, Ptr: &v.Tags},
		{Name: "Steps", Tags: `parquet:"steps"`, Ptr: &v.Steps},
		{Name: "Times", Tags: `parquet:"times,list"`, Ptr: &v.Times},
	}
}

func (v *EventSoa) Cap() int {
	return min(
		len(v.Name),
		len(v.Tags),
		len(v.Steps),
		len(v.Times),
	)
}

func (v *EventSoa) ToAos() iter.Seq[Event] {
	return func(yield func(Event) bool) {
		for i := range v.Cap() {
			item := Event{
				Name:  v.Name[i],
				Tags:  v.Tags[i],
				Steps: v.Steps[i],
				Times: v.Times[i],
			}
			if !yield(item) {
				return
			}
		}
	}
}

func (v *EventSoa) Delete(i int) {
	v.Name[i] = v.Name[len(v.Name)-1]
	v.Name = v.Name[:len(v.Name)-1]
	v.Tags[i] = v.Tags[len(v.Tags)-1]
	v.Tags = v.Tags[:len(v.Tags)-1]
	v.Steps[i] = v.Steps[len(v.Steps)-1]
	v.Steps = v.Steps[:len(v.Steps)-1]
	v.Times[i] = v.Times[len(v.Times)-1]
	v.Times = v.Times[:len(v.Times)-1]
}

func (v *EventSoa) FromAos(aos []Event) *EventSoa {
	v.Allocate(len(aos))
	for i, item := range aos {
		v.Name[i] = item.Name
		v.Tags[i] = item.Tags
		v.Steps[i] = item.Steps
		v.Times[i] = item.Times
	}
	return v
}
