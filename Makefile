build: common_payload_data.go transformations.go
	go build -o sjr *.go

update-jdxd:
	go get -u github.com/whacked/jdxd@main

schemas/%.schema.json: generators/%.schema.jsonnet
	jsonnet $< | jq -S | tee $@

schemas: schemas/CommonPayloadData.schema.json \
	schemas/Transformations.schema.json \
	schemas/CliCommands.schema.json \
	schemas/OpenApi.schema.json

common_payload_data.go: schemas/CommonPayloadData.schema.json
	go-jsonschema --tags json -t -p main $< | \
		sed '/Disabled/! s/\*int/\*int64/g' | \
		tee $@

transformations.go: schemas/Transformations.schema.json
	go-jsonschema --tags json -t -p main $< | \
		sed 's/\*int/\*int64/g' | \
		tee $@

cli_commands.go: schemas/CliCommands.schema.json
	go-jsonschema --tags json -t -p main $< | tee $@
	# generate constants
	cat $< | \
		jq -r '.properties | to_entries | map("const CliConstants_" + (.key | gsub("[^a-zA-Z0-9]"; "_") | ascii_upcase) + " = \"" + .key + "\"") | join("\n")' | \
		tee -a $@

openapi_interface.go: schemas/OpenApi.schema.json
	oapi-codegen -generate types,server -package main $< > $@
