// Generated code by coliosoa. DO NOT EDIT

package clickhouse

import (
	"github.com/davidkleiven/colio/colio"
	"iter"
	"time"
)

type PersonSoa struct {
	Name        []string    `ch:"name"`
	Age         []int64     `ch:"age"`
	NumChildren []int32     `ch:"num_children"`
	Height      []float64   `ch:"height"`
	Weight      []float32   `ch:"weight"`
	CreatedAt   []time.Time `ch:"created_at"`
}

func (v *PersonSoa) Allocate(n int) {
	v.Name = make([]string, n)
	v.Age = make([]int64, n)
	v.NumChildren = make([]int32, n)
	v.Height = make([]float64, n)
	v.Weight = make([]float32, n)
	v.CreatedAt = make([]time.Time, n)
}

func (v *PersonSoa) Columns() []colio.Column {
	return []colio.Column{
		{Name: "Name", Tags: `ch:"name"`, Ptr: &v.Name},
		{Name: "Age", Tags: `ch:"age"`, Ptr: &v.Age},
		{Name: "NumChildren", Tags: `ch:"num_children"`, Ptr: &v.NumChildren},
		{Name: "Height", Tags: `ch:"height"`, Ptr: &v.Height},
		{Name: "Weight", Tags: `ch:"weight"`, Ptr: &v.Weight},
		{Name: "CreatedAt", Tags: `ch:"created_at"`, Ptr: &v.CreatedAt},
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

func (v *PersonSoa) At(i int) Person {
	return Person{
		Name:        v.Name[i],
		Age:         v.Age[i],
		NumChildren: v.NumChildren[i],
		Height:      v.Height[i],
		Weight:      v.Weight[i],
		CreatedAt:   v.CreatedAt[i],
	}
}

type EventSoa struct {
	Name  []string      `ch:"name"`
	Tags  [][]int       `ch:"tags"`
	Names [][]string    `ch:"names"`
	Steps [][]float64   `ch:"steps"`
	Times [][]time.Time `ch:"times"`
}

func (v *EventSoa) Allocate(n int) {
	v.Name = make([]string, n)
	v.Tags = make([][]int, n)
	v.Names = make([][]string, n)
	v.Steps = make([][]float64, n)
	v.Times = make([][]time.Time, n)
}

func (v *EventSoa) Columns() []colio.Column {
	return []colio.Column{
		{Name: "Name", Tags: `ch:"name"`, Ptr: &v.Name},
		{Name: "Tags", Tags: `ch:"tags"`, Ptr: &v.Tags},
		{Name: "Names", Tags: `ch:"names"`, Ptr: &v.Names},
		{Name: "Steps", Tags: `ch:"steps"`, Ptr: &v.Steps},
		{Name: "Times", Tags: `ch:"times"`, Ptr: &v.Times},
	}
}

func (v *EventSoa) Cap() int {
	return min(
		len(v.Name),
		len(v.Tags),
		len(v.Names),
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
				Names: v.Names[i],
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
	v.Names[i] = v.Names[len(v.Names)-1]
	v.Names = v.Names[:len(v.Names)-1]
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
		v.Names[i] = item.Names
		v.Steps[i] = item.Steps
		v.Times[i] = item.Times
	}
	return v
}

func (v *EventSoa) At(i int) Event {
	return Event{
		Name:  v.Name[i],
		Tags:  v.Tags[i],
		Names: v.Names[i],
		Steps: v.Steps[i],
		Times: v.Times[i],
	}
}
