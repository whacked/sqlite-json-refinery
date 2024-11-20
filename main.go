package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "embed"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/whacked/jdxd/pkg/jdxd"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schemas/CliCommands.schema.json
var cliCommandsJson string

func initializeDatabase(db *sql.DB) error {
	_, err := db.Exec(schemaSql)
	return err
}

func insertCommonPayloadData(db *sql.DB, cpd *CommonPayloadData) error {
	_, err := db.Exec(`
		INSERT INTO CommonPayloadData (time, time_added, disabled, device, topic, payload_type, payload)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, cpd.Time, cpd.TimeAdded, cpd.Disabled, cpd.Device, cpd.Topic, cpd.PayloadType, cpd.Payload)
	return err
}

func checkIfTransformationExists(db *sql.DB, inputId int64, transformerId int64) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM Transformations
		WHERE input_id = ? AND transformer_id = ?
	`, inputId, transformerId).Scan(&count)
	return count > 0, err
}

func insertTransformation(db *sql.DB, transformation *Transformations) error {
	fmt.Printf("Inserting transformation with input data id: %d, input schema id: %d, transformer id: %d, output schema id: %d, output id: %d, time executed: %f\n",
		*transformation.InputId,
		*transformation.InputSchemaId,
		*transformation.TransformerId,
		*transformation.OutputSchemaId,
		*transformation.OutputId,
		*transformation.TimeExecuted,
	)
	_, err := db.Exec(`
		INSERT INTO Transformations (input_id, input_schema_id, transformer_id, output_schema_id, output_id, time_executed)
		VALUES (?, ?, ?, ?, ?, ?)
	`, transformation.InputId, transformation.InputSchemaId, transformation.TransformerId, transformation.OutputSchemaId, transformation.OutputId, transformation.TimeExecuted)
	return err
}

func addSchemaPayload(db *sql.DB, content string) error {
	var jsonData interface{}
	fmt.Println("Adding schema payload", content)
	if err := json.Unmarshal([]byte(content), &jsonData); err != nil {
		return err
	}
	canonicalJSON, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	now := float64(time.Now().Unix())
	device := "cli"
	topic := "schema"
	payloadType := CommonPayloadDataPayloadType("schema")
	payload := string(canonicalJSON)
	cpd := &CommonPayloadData{
		Time:        &now,
		TimeAdded:   &now,
		Device:      &device,
		Topic:       &topic,
		PayloadType: &payloadType,
		Payload:     &payload,
	}
	return insertCommonPayloadData(db, cpd)
}

func addTransformerPayload(db *sql.DB, content string) error {
	now := float64(time.Now().Unix())
	device := "cli"
	topic := "transformer"
	payloadType := CommonPayloadDataPayloadType("transformer")
	cpd := &CommonPayloadData{
		Time:        &now,
		TimeAdded:   &now,
		Device:      &device,
		Topic:       &topic,
		PayloadType: &payloadType,
		Payload:     &content,
	}
	return insertCommonPayloadData(db, cpd)
}

func getSummaryStats(db *sql.DB) (map[string]int, error) {
	fmt.Println("Getting summary stats from database", db)

	// print total rows
	rows1, err := db.Query("SELECT COUNT(*) FROM CommonPayloadData")
	if err != nil {
		fmt.Println("Error getting total rows", err)
		return nil, err
	}
	defer rows1.Close()
	var totalRows int
	for rows1.Next() {
		if err := rows1.Scan(&totalRows); err != nil {
			fmt.Println("Error getting total rows", err)
			return nil, err
		}
	}

	fmt.Println("Total rows:", totalRows)

	stats := make(map[string]int)
	rows, err := db.Query("SELECT payload_type, COUNT(*) FROM CommonPayloadData GROUP BY payload_type")
	if err != nil {
		fmt.Println("Error getting summary stats", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var payloadType string
		var count int
		if err := rows.Scan(&payloadType, &count); err != nil {
			return nil, err
		}
		stats[payloadType] = count
	}
	return stats, nil
}

func listPayloads(db *sql.DB, payloadType string) error {
	fmt.Println("Listing payloads of type", payloadType)
	rows, err := db.Query("SELECT id, payload FROM CommonPayloadData WHERE payload_type = ?", payloadType)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var payload string
		if err := rows.Scan(&id, &payload); err != nil {
			return err
		}
		if payloadType == "schema" || payloadType == "raw" || payloadType == "derived" {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, []byte(payload), "", "  "); err != nil {
				return err
			}
			fmt.Printf("[%d] %s\n", id, prettyJSON.String())
		} else {
			fmt.Printf("[%d] %s\n", id, payload)
		}
	}
	return nil
}

func applyTransform(db *sql.DB, inputID int64, inputSchemaID int64, transformerID int64, outputSchemaID int64) error {
	var inputData, inputSchema, transformer, outputSchema CommonPayloadData
	records := []struct {
		id     int64
		record *CommonPayloadData
	}{
		{inputID, &inputData},
		{inputSchemaID, &inputSchema},
		{transformerID, &transformer},
		{outputSchemaID, &outputSchema},
	}

	for _, r := range records {
		if err := db.QueryRow("SELECT payload, payload_type, device, topic FROM CommonPayloadData WHERE id = ?", r.id).Scan(&r.record.Payload, &r.record.PayloadType, &r.record.Device, &r.record.Topic); err != nil {
			return fmt.Errorf("error loading record %d: %v", r.id, err)
		}
	}

	fmt.Println("Checking if transformation exists")
	checkTransformation, err := checkIfTransformationExists(db, inputID, transformerID)
	if err != nil {
		return fmt.Errorf("error checking if transformation exists: %v", err)
	}
	if checkTransformation {
		return fmt.Errorf("transformation already exists")
	}
	fmt.Println("Transformation does not exist")

	inputValidator := jdxd.MakeRecordValidatorFromJsonString("input", *inputSchema.Payload)
	outputValidator := jdxd.MakeRecordValidatorFromJsonString("output", *outputSchema.Payload)

	recordTransformer := jdxd.MakeRecordTransformer(*transformer.Payload, "jsonata", "")

	var inputDataInterface interface{}
	if err := json.Unmarshal([]byte(*inputData.Payload), &inputDataInterface); err != nil {
		return fmt.Errorf("error unmarshalling input data: %v", err)
	}

	output := jdxd.TransformRecord(&inputDataInterface, inputValidator, outputValidator, recordTransformer)

	prettyOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("error pretty printing output: %v", err)
	}
	fmt.Println(string(prettyOutput))

	jsonOutput, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("error marshalling output: %v", err)
	}

	now := float64(time.Now().Unix())
	device := "cli"
	payloadType := CommonPayloadDataPayloadType("derived")
	payload := string(jsonOutput)
	newCommonPayloadData := &CommonPayloadData{
		Time:        &now,
		TimeAdded:   &now,
		Device:      &device,
		Topic:       inputData.Topic,
		PayloadType: &payloadType,
		Payload:     &payload,
	}

	insertCommonPayloadData(db, newCommonPayloadData)

	// get the id of the new common payload data
	var newCommonPayloadDataID int64
	if err := db.QueryRow("SELECT id FROM CommonPayloadData WHERE payload = ?", payload).Scan(&newCommonPayloadDataID); err != nil {
		return fmt.Errorf("error getting id of new common payload data: %v", err)
	}

	// attempt to insert the Transformations record first
	transformation := &Transformations{
		InputId:        &inputID,
		InputSchemaId:  &inputSchemaID,
		TransformerId:  &transformerID,
		OutputSchemaId: &outputSchemaID,
		OutputId:       &newCommonPayloadDataID,
		TimeExecuted:   &now,
	}
	fmt.Printf("Inserting transformation: %+v\n", transformation)
	if err := insertTransformation(db, transformation); err != nil {
		return fmt.Errorf("error inserting transformations: %v", err)
	}
	fmt.Println("Transformation inserted")

	return nil
}

type Config struct {
	Database            string
	SubCommand          string
	MQTTHost            string
	MQTTPort            int
	Content             string
	Type                string
	Spec                string
	InputDataID         int
	InputSchemaID       int
	OutputTransformerID int
	OutputSchemaID      int
	tempThing           *CliCommands
}

// Create a map from command strings to struct fields
// returns the json tag ==> type mapping
func buildCommandMapping(commands *CliCommands) map[string]interface{} {
	mapping := make(map[string]interface{})
	val := reflect.ValueOf(commands).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := field.Tag.Get("json")
		// fmt.Println("(processing field)", i, field, jsonTag)
		if jsonTag != "" {
			// Strip `omitempty` or other modifiers from the tag
			jsonKey := jsonTag
			if commaIdx := strings.Index(jsonTag, ","); commaIdx != -1 {
				jsonKey = jsonTag[:commaIdx]
			}
			fieldValue := val.Field(i).Interface()
			// Get the type to instantiate - for pointers, use the element type
			targetType := reflect.TypeOf(fieldValue)
			if targetType.Kind() == reflect.Ptr {
				targetType = targetType.Elem()
			}
			instance := reflect.New(targetType).Interface()
			// fmt.Printf("ADDING INSTANCE: %s -> %v\n", jsonKey, instance)
			mapping[jsonKey] = instance
		}
	}

	return mapping
}

func setupFlags(commandMapping map[string]interface{}, cliCommandsJsonData map[string]interface{}, attribute string, flagSet *flag.FlagSet) {

	subCommandMapping := commandMapping[attribute]
	val := reflect.ValueOf(subCommandMapping).Elem()
	typ := val.Type()

	// fmt.Printf("subCommandMapping for [%s]: %+v\n", attribute, typ)

	attributeSchema := cliCommandsJsonData["properties"].(map[string]interface{})[attribute].(map[string]interface{})
	// bail if the attribute schema is not type == object
	if attributeSchema["type"] != "object" {

		switch attributeSchema["type"] {
		case "boolean":
			defaultValue := true // true because user passed in the flag
			flagSet.BoolVar(&defaultValue, attribute, defaultValue, "")
			return
		}
		fmt.Printf("Attribute %s is not an object; it is %s\n", attribute, attributeSchema["type"])

		return
	}

	// check if it has the attribute "properties"
	if _, ok := attributeSchema["properties"]; !ok {
		fmt.Printf("No properties found for attribute: %s\n", attribute)
		return
	}

	attributeSchemaProperties := attributeSchema["properties"].(map[string]interface{})
	if len(attributeSchemaProperties) == 0 {
		fmt.Printf("No properties found for attribute: %s\n", attribute)
		return
	}

	for i := 0; i < val.NumField(); i++ {

		field := typ.Field(i)
		jsonTag := field.Tag.Get("json")
		if commaIdx := strings.Index(jsonTag, ","); commaIdx != -1 {
			jsonTag = jsonTag[:commaIdx]
		}
		if jsonTag == "" {
			continue
		}

		subPropertySchema := attributeSchemaProperties[jsonTag].(map[string]interface{})
		maybeDescription := subPropertySchema["description"]
		description := ""
		if maybeDescription != nil {
			description = maybeDescription.(string)
		}

		switch field.Type.Kind() {
		case reflect.String:
			defaultValue := subPropertySchema["default"]
			flagSet.StringVar(val.Field(i).Addr().Interface().(*string), jsonTag, defaultValue.(string), description)
		case reflect.Int:
			defaultValue := subPropertySchema["default"]
			// coerce the default value to an int; the json read-in yields a float64
			defaultInt := int(defaultValue.(float64))
			flagSet.IntVar(val.Field(i).Addr().Interface().(*int), jsonTag, defaultInt, description)
		case reflect.Bool:
			defaultValue := subPropertySchema["default"]
			flagSet.BoolVar(val.Field(i).Addr().Interface().(*bool), jsonTag, defaultValue.(bool), description)
		case reflect.Ptr:
			// For pointer fields, we need to get the pointer to the pointer field
			// This handles cases like *string in the struct
			defaultValue := ""
			if def, ok := subPropertySchema["default"]; ok {
				defaultValue = def.(string)
			}
			// Get pointer to the *string field
			ptrToPtr := val.Field(i).Addr().Interface().(**string)
			// Create storage for the string value
			*ptrToPtr = new(string)
			// Pass the pointer to the string value to flagSet
			flagSet.StringVar(*ptrToPtr, jsonTag, defaultValue, description)
		default:
			fmt.Printf("-- Unsupported flag type for field %s, type %s\n", field.Name, field.Type.Kind())
		}
	}
}

func parseCommandLine() Config {
	commands := &CliCommands{}
	commandMapping := buildCommandMapping(commands)

	var cliCommandsJsonData map[string]interface{}
	json.Unmarshal([]byte(cliCommandsJson), &cliCommandsJsonData)
	// fmt.Printf("=== cliCommandsJsonData: %+v\n", cliCommandsJsonData)

	// Create a map of subcommands to their FlagSets
	subCommandFlags := make(map[string]*flag.FlagSet)
	// Initialize FlagSets for each subcommand
	for cmdName := range commandMapping {
		subCommandFlags[cmdName] = flag.NewFlagSet(cmdName, flag.ExitOnError)

		if cmdName == "serve-jsonl" {
			fmt.Printf("serve-jsonl flags: %+v\n", subCommandFlags[cmdName])
			// Create storage for the string value and assign it to the pointer field
			commands.ServeJsonl = new(CliCommandsServeJsonl)
			commands.ServeJsonl.Source = new(string)
			subCommandFlags[cmdName].StringVar(commands.ServeJsonl.Source, "source", "", "path to JSONL file to serve")
		} else {
			setupFlags(commandMapping, cliCommandsJsonData, cmdName, subCommandFlags[cmdName])
		}
	}

	/*
		// Set up flags for each subcommand
		if connectCmd, ok := subcommandFlags["connect"]; ok {
			fmt.Printf("mapping: %+v\n", commandMapping)
			setupFlags(commandMapping, cliCommandsJsonData, "connect", connectCmd)
		}

		if addSchemaCmd, ok := subcommandFlags["add-schema"]; ok {
			setupFlags(commandMapping, cliCommandsJsonData, "add-schema", addSchemaCmd)
			// addSchemaCmd.StringVar(&cfg.Content, "content", "", "File path, raw JSON, or '-' for stdin")
		}

		if addTransformerCmd, ok := subcommandFlags["add-transformer"]; ok {
			setupFlags(commandMapping, cliCommandsJsonData, "add-transformer", addTransformerCmd)
			// addTransformerCmd.StringVar(&cfg.Content, "content", "", "File path, raw string, or '-' for stdin")
		}

		if listPayloadsCmd, ok := subcommandFlags["list-payloads"]; ok {
			setupFlags(commandMapping, cliCommandsJsonData, "list-payloads", listPayloadsCmd)
			// listPayloadsCmd.StringVar(&cfg.Type, "type", "raw", "Payload type (schema, transformer, raw)")
		}

		if applyTransformCmd, ok := subcommandFlags["apply-transform"]; ok {
			setupFlags(commandMapping, cliCommandsJsonData, "apply-transform", applyTransformCmd)
			// applyTransformCmd.StringVar(&cfg.Spec, "spec", "", "Transform specification in the shape of input:inputSchemaID/outputTransformerID:outputSchemaID")
		}
	*/

	cfg := Config{}
	cfg.tempThing = commands

	// Set up the database flag
	dbDefault := ":memory:"
	if envDB := os.Getenv("DATABASE_PATH"); envDB != "" {
		dbDefault = envDB
	}
	flag.StringVar(&cfg.Database, "database", dbDefault, "Database path")

	// Add a custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [global options] <subcommand> [subcommand options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Global options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
		subCommandKeys := make([]string, 0, len(commandMapping))
		for cmd := range commandMapping {
			subCommandKeys = append(subCommandKeys, cmd)
		}
		sort.Strings(subCommandKeys)
		for _, cmd := range subCommandKeys {
			fmt.Fprintf(os.Stderr, "  %s\n", cmd)
		}
		fmt.Fprintf(os.Stderr, "\nUse '%s <subcommand> -h' for more information about a subcommand.\n", os.Args[0])
	}

	// Parse the global flags
	flag.Parse()

	// Print the database path
	fmt.Printf("Using database: %s\n", cfg.Database)

	// Check if no subcommand was provided or -h was used
	if len(flag.Args()) < 1 || (len(os.Args) > 1 && os.Args[1] == "-h") {
		flag.Usage()
		os.Exit(1)
	}
	/*


			// debug print all args detected by flag
			fmt.Printf("Args: %v\n", flag.Args())
			// Parse the subcommand's flags
			if subCmdParser, ok := subcommandFlags[subCommand]; ok {
				subCmdParser.Parse(flag.Args()[1:])
			} else {
				fmt.Printf("Unknown subcommand: %s\n", subCommand)
				flag.Usage()
				os.Exit(1)
			}

			// Set up the database flag
			dbDefault := ":memory:"
			if envDB := os.Getenv("DATABASE_PATH"); envDB != "" {
				dbDefault = envDB
			}
			flag.StringVar(&cfg.Database, "database", dbDefault, "Database path")

			fmt.Println("OK?")
			flag.Usage()

			os.Exit(0)

			// Add a custom usage function
			flag.Usage = func() {
				fmt.Fprintf(os.Stderr, "Usage: %s [global options] <subcommand> [subcommand options]\n\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "Global options:\n")
				flag.PrintDefaults()
				fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
				for _, cmd := range subcommands {
					fmt.Fprintf(os.Stderr, "  %s\n", cmd)
				}
				fmt.Fprintf(os.Stderr, "\nUse '%s <subcommand> -h' for more information about a subcommand.\n", os.Args[0])
			}

			// Parse the global flags
			flag.Parse()

			// Print the database path
			fmt.Printf("Using database: %s\n", cfg.Database)

			// Check if no subcommand was provided or -h was used
			if len(flag.Args()) < 1 || (len(os.Args) > 1 && os.Args[1] == "-h") {
				flag.Usage()
				fmt.Println("OK???")
				os.Exit(1)
			}

					// Example: Look up a command
					command := "add-schema"
					if field, found := commandMapping[command]; found {
						fmt.Printf("Command %s matches struct field %+v\n", command, field)
					} else {
						fmt.Printf("Command %s not found\n", command)
					}


				cfg.SubCommand = flag.Arg(0)

		// Define subcommands
		connectCmd := flag.NewFlagSet("connect", flag.ExitOnError)
		addSchemaCmd := flag.NewFlagSet("add-schema", flag.ExitOnError)
		addTransformerCmd := flag.NewFlagSet("add-transformer", flag.ExitOnError)
		summaryCmd := flag.NewFlagSet("summary", flag.ExitOnError)
		listPayloadsCmd := flag.NewFlagSet("list-payloads", flag.ExitOnError)
		listSchemasCmd := flag.NewFlagSet("list-schemas", flag.ExitOnError)
		listTransformersCmd := flag.NewFlagSet("list-transformers", flag.ExitOnError)
		applyTransformCmd := flag.NewFlagSet("apply-transform", flag.ExitOnError)

		// connect subcommand flags
		connectCmd.StringVar(&cfg.MQTTHost, "mqtt-host", "localhost", "MQTT broker hostname")
		connectCmd.IntVar(&cfg.MQTTPort, "mqtt-port", 1883, "MQTT broker port")

		// add-schema and add-transformer subcommand flags
		addSchemaCmd.StringVar(&cfg.Content, "content", "", "File path, raw JSON, or '-' for stdin")
		addTransformerCmd.StringVar(&cfg.Content, "content", "", "File path, raw string, or '-' for stdin")

		// list-payloads subcommand flag
		listPayloadsCmd.StringVar(&cfg.Type, "type", "raw", "Payload type (schema, transformer, raw)")

		// apply-transform subcommand flag
		applyTransformCmd.StringVar(&cfg.Spec, "spec", "", "Transform specification in the shape of input:inputSchemaID/outputTransformerID:outputSchemaID")

		switch cfg.SubCommand {
		case "connect":
			connectCmd.Parse(flag.Args()[1:])
		case "add-schema":
			addSchemaCmd.Parse(flag.Args()[1:])
		case "add-transformer":
			addTransformerCmd.Parse(flag.Args()[1:])
		case "summary":
			summaryCmd.Parse(flag.Args()[1:])
		case "list-payloads":
			listPayloadsCmd.Parse(flag.Args()[1:])
		case "list-schemas":
			listSchemasCmd.Parse(flag.Args()[1:])
		case "list-transformers":
			listTransformersCmd.Parse(flag.Args()[1:])
		case "apply-transform":
			applyTransformCmd.Parse(flag.Args()[1:])
		default:
			fmt.Printf("Unknown subcommand: %s\n", cfg.SubCommand)
			flag.Usage()
			os.Exit(1)
		}
	*/

	cfg.SubCommand = flag.Arg(0)

	if subCommandHandler, ok := subCommandFlags[cfg.SubCommand]; ok {
		fmt.Println("subCommandHandler", cfg.SubCommand, subCommandHandler)
		fmt.Println("parsing", flag.Args()[1:])
		subCommandHandler.Parse(flag.Args()[1:])
	} else {
		fmt.Printf("Unknown subcommand: %s\n", cfg.SubCommand)
		flag.Usage()
		os.Exit(1)
	}

	// Handle stdin for add-schema and add-transformer
	if cfg.SubCommand == "add-schema" || cfg.SubCommand == "add-transformer" {
		if cfg.Content == "" || cfg.Content == "-" {
			fmt.Println("Reading from stdin...")
			// Read from stdin
			content, err := readStdin()
			if err != nil {
				fmt.Println("Error reading from stdin:", err)
				os.Exit(1)
			}
			cfg.Content = content
		}
	}

	// Parse the spec for apply-transform
	if cfg.SubCommand == "apply-transform" {
		parseSpec(&cfg)
	}

	return cfg
}

func readStdin() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	var content strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		content.WriteString(line)
	}

	return content.String(), nil
}

func parseSpec(cfg *Config) {
	re := regexp.MustCompile(`^(\d+):?(\d+)?/(\d+):?(\d+)?$`)
	matches := re.FindStringSubmatch(cfg.Spec)

	if len(matches) != 5 {
		fmt.Println("Invalid spec format")
		os.Exit(1)
	}

	cfg.InputDataID, _ = strconv.Atoi(matches[1])
	cfg.InputSchemaID, _ = strconv.Atoi(matches[2])
	cfg.OutputTransformerID, _ = strconv.Atoi(matches[3])
	cfg.OutputSchemaID, _ = strconv.Atoi(matches[4])
}

func main() {
	jdxd.SetDebugLevelFromEnvironment()

	config := parseCommandLine()
	fmt.Printf("Parsed configuration: %+v\n", config)
	// pretty print the parsed configuration
	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(b))

	db, err := sql.Open("sqlite3", config.Database)
	if err != nil {
		log.Fatal(err)
	}
	if err := initializeDatabase(db); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	switch config.SubCommand {
	case CliConstants_ADD_SCHEMA:
		if err := addSchemaPayload(db, config.Content); err != nil {
			log.Fatal(err)
		}

	case CliConstants_ADD_TRANSFORMER:
		if err := addTransformerPayload(db, config.Content); err != nil {
			log.Fatal(err)
		}

	case CliConstants_SUMMARY:
		stats, err := getSummaryStats(db)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Summary stats:")
		for payloadType, count := range stats {
			fmt.Printf("%s: %d\n", payloadType, count)
		}

	case CliConstants_LIST_PAYLOADS:
		if err := listPayloads(db, config.Type); err != nil {
			log.Fatal(err)
		}

	case CliConstants_LIST_SCHEMAS:
		if err := listPayloads(db, "schema"); err != nil {
			log.Fatal(err)
		}

	case CliConstants_LIST_TRANSFORMERS:
		if err := listPayloads(db, "transformer"); err != nil {
			log.Fatal(err)
		}

	case CliConstants_APPLY_TRANSFORM:
		if err := applyTransform(db, int64(config.InputDataID), int64(config.InputSchemaID), int64(config.OutputTransformerID), int64(config.OutputSchemaID)); err != nil {
			log.Fatal(err)
		}

	case CliConstants_CONNECT:
		opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", config.MQTTHost, config.MQTTPort))
		hostname, _ := os.Hostname()
		opts.SetClientID(fmt.Sprintf("mqtt-dumper-%s", hostname))

		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}

		if token := client.Subscribe("#", 0, func(client mqtt.Client, msg mqtt.Message) {
			now := float64(time.Now().Unix())
			device := "mqtt-dumper"
			topic := msg.Topic()
			payloadType := CommonPayloadDataPayloadType("raw")
			payload := string(msg.Payload())
			cpd := &CommonPayloadData{
				Time:        &now,
				TimeAdded:   &now,
				Device:      &device,
				Topic:       &topic,
				PayloadType: &payloadType,
				Payload:     &payload,
			}

			if err := insertCommonPayloadData(db, cpd); err != nil {
				log.Printf("Error inserting data: %v", err)
			} else {
				log.Printf("Inserted: %+v", cpd)
			}
		}); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}

		log.Println("Connected to MQTT broker. Press CTRL+C to exit.")
		select {}

	case CliConstants_SERVE_JSONL:
		ServeJsonl(&config)

	default:
		log.Fatalf("Unknown subcommand: %s", config.SubCommand)
	}
}
