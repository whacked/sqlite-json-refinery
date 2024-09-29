package main

import _ "embed"

//go:embed db/schema.sql
var schemaSql string
