{
  openapi: '3.0.0',
  info: {
    version: '1.0.0',
    title: 'JSONL Server',
    license: {
      name: 'MIT',
    },
  },
  servers: [
    {
      //      "url": "http://petstore.swagger.io/v1"
    },
  ],
  paths: {
    '/count': {
      get: {
        summary: 'Count all records',
        operationId: 'countRecords',
        tags: [],
        parameters: [
        ],
        responses: {
          '200': {
            description: 'total number of records',
            schema: {
              type: 'integer',
            },
            content: {
              'application/json': {},
            },
          },
          default: {
            description: 'unexpected error',
            content: {
              'application/json': {
                schema: {
                  '$ref': '#/components/schemas/Error',
                },
              },
            },
          },
        },
      },
    },
    '/records': {
      get: {
        summary: 'List paginated records',
        operationId: 'listRecords',
        tags: [],
        parameters: [
          {
            name: 'limit',
            'in': 'query',
            description: 'Number of records to return',
            required: false,
            schema: {
              type: 'integer',
              minimum: 1,
            },
          },
          {
            name: 'offset',
            'in': 'query',
            description: 'Number of records to skip',
            required: false,
            schema: {
              type: 'integer',
              minimum: 0,
            },
          },
        ],
        responses: {
          '200': {
            description: 'paginated list of records',
            content: {
              'application/json': {
                schema: {
                  '$ref': '#/components/schemas/Records',
                },
              },
            },
          },
        },
      },
    },
    '/records/{index}': {
      get: {
        summary: 'Info for a specific record',
        operationId: 'showRecordById',
        tags: [],
        parameters: [
          {
            name: 'index',
            'in': 'path',
            required: true,
            description: 'The index of the record to retrieve',
            schema: {
              type: 'integer',
            },
          },
        ],
        responses: {
          '200': {
            description: 'Expected response to a valid request',
            content: {
              'application/json': {
                schema: {
                  '$ref': '#/components/schemas/Record',
                },
              },
            },
          },
          default: {
            description: 'unexpected error',
            content: {
              'application/json': {
                schema: {
                  '$ref': '#/components/schemas/Error',
                },
              },
            },
          },
        },
      },
    },
  },
  components: {
    schemas: {
      Record: {
        type: 'object',
      },
      Records: {
        type: 'array',
        // maxItems: 100,
        items: {
          '$ref': '#/components/schemas/Record',
        },
      },
      Error: {
        type: 'object',
        required: ['code', 'message'],
        properties: {
          code: {
            type: 'integer',
            format: 'int32',
          },
          message: {
            type: 'string',
          },
        },
      },
    },
  },
}
