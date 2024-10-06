{
  title: 'CommonPayloadData',
  type: 'object',
  description: 'main table for storing and tracking data',
  properties: {
    id: {
      description: 'auto-generated primary key',
      type: 'integer',
    },
    time: {
      type: 'number',
    },
    time_added: {
      type: 'number',
    },
    disabled: {
      description: 'marks data to be skipped',
      type: 'integer',
    },
    device: {
      description: 'name of the device from which the data is collected',
      type: 'string',
      examples: [
        'rpizero-2',
        'cli',
        'app@2.1',
      ],
    },
    topic: {
      description: 'maps directly to the MQTT topic, but is otherwise an arbitrary "project"/"group"-like identifier',
      type: 'string',
    },
    payload_type: {
      description: 'for internal organization; raw: raw payload data; schema: json schema; transformer: jsonata (later also jsonnet) data transformer; derived: post-transform data',
      type: 'string',
      enum: ['raw', 'schema', 'transformer', 'derived'],
    },
    payload: {
      description: 'raw data: json payloads, json schema content, transformer code',
      type: 'string',
    },
  },
}
