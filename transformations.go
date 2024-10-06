// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package main

// tracks the lineage of transformed data, all joining to the CommonPayloadData
// table
type Transformations struct {
	// auto-generated primary key
	Id *int64 `json:"id,omitempty"`

	// InputId corresponds to the JSON schema field "input_id".
	InputId *int64 `json:"input_id,omitempty"`

	// InputSchemaId corresponds to the JSON schema field "input_schema_id".
	InputSchemaId *int64 `json:"input_schema_id,omitempty"`

	// OutputId corresponds to the JSON schema field "output_id".
	OutputId *int64 `json:"output_id,omitempty"`

	// OutputSchemaId corresponds to the JSON schema field "output_schema_id".
	OutputSchemaId *int64 `json:"output_schema_id,omitempty"`

	// TimeExecuted corresponds to the JSON schema field "time_executed".
	TimeExecuted *float64 `json:"time_executed,omitempty"`

	// TransformerId corresponds to the JSON schema field "transformer_id".
	TransformerId *int64 `json:"transformer_id,omitempty"`
}
