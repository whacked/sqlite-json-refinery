{
  title: 'Transformations',
  type: 'object',
  description: 'tracks the lineage of transformed data, all joining to the CommonPayloadData table',
  properties: {
    id: {
      description: 'auto-generated primary key',
      type: 'integer',
    },
    input_id: {
      type: 'integer',
    },
    input_schema_id: {
      type: 'integer',
    },
    transformer_id: {
      type: 'integer',
    },
    output_id: {
      type: 'integer',
    },
    output_schema_id: {
      type: 'integer',
    },
    time_executed: {
      type: 'number',
    },
  },
}
