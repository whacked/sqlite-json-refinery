CREATE TABLE IF NOT EXISTS CommonPayloadData (
    id INTEGER PRIMARY KEY,
    time INTEGER NOT NULL,
    time_added INTEGER NOT NULL,
    disabled INTEGER DEFAULT 0,
    device TEXT NOT NULL,
    topic TEXT NOT NULL,
    payload_type TEXT NOT NULL DEFAULT 'raw',  -- raw, derived, schema, transformer
    payload TEXT NOT NULL
    -- UNIQUE(device, topic, payload)
);

CREATE TABLE IF NOT EXISTS Transformations (
    id INTEGER PRIMARY KEY,
    input_id INTEGER NOT NULL,
    input_schema_id INTEGER,
    transformer_id INTEGER NOT NULL,
    output_id INTEGER NOT NULL,
    output_schema_id INTEGER,
    time_executed INTEGER NOT NULL,
    FOREIGN KEY (input_id) REFERENCES CommonPayloadData(id),
    FOREIGN KEY (input_schema_id) REFERENCES CommonPayloadData(id),
    FOREIGN KEY (transformer_id) REFERENCES CommonPayloadData(id),
    FOREIGN KEY (output_id) REFERENCES CommonPayloadData(id),
    FOREIGN KEY (output_schema_id) REFERENCES CommonPayloadData(id),
    UNIQUE (input_id, transformer_id)
);
