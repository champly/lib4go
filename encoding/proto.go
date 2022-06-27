package encoding

import (
	"github.com/ghodss/yaml"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// JSON2Struct json string transform structpb.Struct
func JSON2Struct(str string) (*structpb.Struct, error) {
	// "github.com/gogo/protobuf/types"
	// result := &types.Struct{}
	result := &structpb.Struct{}

	m := protojson.UnmarshalOptions{}
	err := m.Unmarshal([]byte(str), result)
	return result, err
}

// YAML2Struct yaml string transform structpb.Struct
func YAML2Struct(str string) (*structpb.Struct, error) {
	b, err := yaml.YAMLToJSON([]byte(str))
	if err != nil {
		return nil, err
	}
	return JSON2Struct(string(b))
}
