update-jdxd:
	go get -u github.com/whacked/jdxd@main

schemas/%.schema.json: generators/%.schema.jsonnet
	jsonnet $< | jq -S | tee $@

schemas: schemas/CommonPayloadData.schema.json schemas/Transformations.schema.json

common_payload_data.go: schemas/CommonPayloadData.schema.json
	go-jsonschema --tags json -t -p main $< | \
		sed '/Disabled/! s/\*int/\*int64/g' | \
		tee $@

transformations.go: schemas/Transformations.schema.json
	go-jsonschema --tags json -t -p main $< | \
		sed 's/\*int/\*int64/g' | \
		tee $@