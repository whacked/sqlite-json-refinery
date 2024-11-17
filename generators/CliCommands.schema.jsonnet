{
  title: 'CliCommands',
  type: 'object',
  description: 'cli commands and parameters',
  properties: {
    /*
    connect: {
      type: 'object',
      properties: {
        'mqtt-host': {
          type: 'string',
          description: 'MQTT host for connection',
          default: 'localhost',
        },
        'mqtt-port': {
          type: 'integer',
          description: 'MQTT port for connection',
          default: 1883,
        },
      },
    },
    'add-schema': {
      type: 'object',
      properties: {
        content: {
          type: 'string',
        },
      },
    },
    'add-transformer': {
      type: 'object',
      properties: {
        content: {
          type: 'string',
        },
      },
    },
    summary: {
      type: 'boolean',
    },
    'list-payloads': {
      type: 'object',
      properties: {
        type: {
          type: 'string',
        },
      },
    },
    'list-schemas': {
      type: 'boolean',
    },
    'list-transformers': {
      type: 'boolean',
    },
    'apply-transform': {
      type: 'object',
      properties: {
        spec: {
          type: 'string',
        },
      },
    },
    */
    'serve-jsonl': {
      type: 'object',
      properties: {
        port: {
          type: 'integer',
          description: 'port to serve JSONL on',
          default: 8080,
        },
        source: {
          type: 'string',
          description: 'path to JSONL file to serve',
        },
      },
    },
  },
}
