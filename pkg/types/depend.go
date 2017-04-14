package types

import (
	"encoding/json"
	"reflect"

	"github.com/twmb/algoimpl/go/graph"
)

// converts a map to a Spec, nil if it cannot be done
func mapToSpec(m map[string]interface{}) *Spec {
	// This is hacky -- generate a string representation
	// and try to parse it as struct
	buff, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	s := Spec{}
	err = json.Unmarshal(buff, &s)
	if err != nil {
		return nil
	}
	if s.Validate() == nil {
		return &s
	}
	return nil
}

// findSpecs parses the bytes and returns a Spec, if the Spec can be parsed
// from the buffer.  Some fields are verified and must be present for the
// buffer to be considered a representation of a Spec.
func findSpecs(v interface{}) []*Spec {

	result := []*Spec{}

	switch v := v.(type) {

	case []*Spec:
		for _, x := range v {
			c := *x
			result = append(result, findSpecs(&c)...)
		}

	case []Spec:
		for _, x := range v {
			c := x
			result = append(result, findSpecs(&c)...)
		}

	case []interface{}:
		for _, x := range v {
			c := x
			result = append(result, findSpecs(c)...)
		}

	case map[string]interface{}:
		// convert to Spec?
		result = append(result, findSpecs(mapToSpec(v))...)

	case *Any:

		if v == nil {
			return result
		}

		spec := Spec{}
		if err := v.Decode(&spec); err == nil {

			if spec.Validate() == nil {
				result = append(result, findSpecs(&spec)...)
				return result
			}

		}

		// now decode as regular struct - map or []interface{}
		var vv interface{}
		if err := v.Decode(&vv); err != nil {
			return nil
		}

		switch vv := vv.(type) {
		case []interface{}:
			for _, x := range vv {
				result = append(result, findSpecs(x)...)
			}
		case map[string]interface{}:
			for _, x := range vv {
				result = append(result, findSpecs(x)...)
			}
		}

	case Spec:
		c := v
		result = append(result, &c)
		result = append(result, findSpecs(c.Properties)...)

	case *Spec:

		if v == nil {
			return result
		}

		c := *v
		result = append(result, &c)
		result = append(result, findSpecs(c.Properties)...)

	default:
		value := reflect.Indirect(reflect.ValueOf(v))
		if value.Type().Kind() == reflect.Struct {
			for i := 0; i < value.NumField(); i++ {
				fv := value.Field(i)
				if fv.IsValid() {
					result = append(result, findSpecs(fv.Interface())...)
				}
			}
		}
	}
	return result
}

// Nested recurses through the Properties of the spec and returns any nested specs.
func (s Spec) Nested() []*Spec {
	if s.Properties == nil {
		return nil
	}
	return findSpecs(s.Properties)
}

type key struct {
	class string
	name  string
}

func indexSpecs(specs []*Spec, g *graph.Graph) map[key]*graph.Node {
	index := map[key]*graph.Node{}
	for _, spec := range specs {

		node := g.MakeNode()
		*(node.Value) = spec

		index[key{class: spec.Class, name: spec.Metadata.Name}] = &node
	}
	return index
}

func indexGet(index map[key]*graph.Node, class, name string) *graph.Node {
	if v, has := index[key{class: class, name: name}]; has {
		return v
	}
	return nil
}

// OrderByDependency returns the given specs in dependency order.
func OrderByDependency(specs []*Spec) ([]*Spec, error) {

	g := graph.New(graph.Directed)
	if g == nil {
		return nil, nil
	}

	index := indexSpecs(specs, g)

	for _, spec := range specs {

		from := indexGet(index, spec.Class, spec.Metadata.Name)
		if from == nil {
			return nil, errNotFound{class: spec.Class, name: spec.Metadata.Name}
		}

		for _, depend := range spec.Depends {

			to := indexGet(index, depend.Class, depend.Name)
			if to == nil {
				return nil, errBadDependency(depend)
			}

			if from == to {

				a := (*from.Value).(*Spec)
				b := (*to.Value).(*Spec)
				return nil, errCircularDependency([]*Spec{a, b})
			}

			if err := g.MakeEdge(*to, *from); err != nil {
				return nil, err
			}
		}
	}

	// cycle detection
	for _, connected := range g.StronglyConnectedComponents() {
		if len(connected) > 1 {

			cycle := []*Spec{}
			for _, n := range connected {
				cycle = append(cycle, (*n.Value).(*Spec))
			}
			return nil, errCircularDependency(cycle)
		}
	}

	ordered := []*Spec{}
	for _, n := range g.TopologicalSort() {
		ordered = append(ordered, (*n.Value).(*Spec))
	}

	return ordered, nil
}
