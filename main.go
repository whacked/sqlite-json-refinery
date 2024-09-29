package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/whacked/jdxd/pkg/jdxd"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/mattn/go-sqlite3"
)

type CommonPayloadData struct {
	ID          int64  `json:"id"`
	Time        int64  `json:"time"`
	TimeAdded   int64  `json:"time_added"`
	Disabled    int    `json:"disabled"`
	Device      string `json:"device"`
	Topic       string `json:"topic"`
	PayloadType string `json:"payload_type"`
	Payload     string `json:"payload"`
}

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
	cpd := &CommonPayloadData{
		Time:        time.Now().Unix(),
		TimeAdded:   time.Now().Unix(),
		Device:      "cli",
		Topic:       "schema",
		PayloadType: "schema",
		Payload:     string(canonicalJSON),
	}
	return insertCommonPayloadData(db, cpd)
}

func addTransformerPayload(db *sql.DB, content string) error {
	cpd := &CommonPayloadData{
		Time:        time.Now().Unix(),
		TimeAdded:   time.Now().Unix(),
		Device:      "cli",
		Topic:       "transformer",
		PayloadType: "transformer",
		Payload:     content,
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
	// Implementation of transform application logic goes here
	// This is a placeholder and would require additional logic to execute the transformer
	return nil
}

func main() {
	jdxd.SetDebugLevelFromEnvironment()
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	databaseFile := ":memory:"
	parseArgsIndex := 1
	// look for a -database flag first
	for i, arg := range os.Args {
		if arg == "-database" {
			databaseFile = os.Args[i+1]
			parseArgsIndex = i + 2
			break
		}
	}

	// Define subcommands
	connectCmd := flag.NewFlagSet("connect", flag.ExitOnError)
	mqttHost := connectCmd.String("mqtt-host", "localhost", "MQTT broker host")
	mqttPort := connectCmd.Int("mqtt-port", 1883, "MQTT broker port")
	clientName := flag.String("client-name", fmt.Sprintf("mqtt-dumper-%s", hostname), "MQTT client name")

	addSchemaCmd := flag.NewFlagSet("add-schema", flag.ExitOnError)
	schemaContent := addSchemaCmd.String("content", "", "Schema content or file path")

	addTransformerCmd := flag.NewFlagSet("add-transformer", flag.ExitOnError)
	transformerContent := addTransformerCmd.String("content", "", "Transformer content or file path")

	summaryCmd := flag.NewFlagSet("summary", flag.ExitOnError)

	listPayloadsCmd := flag.NewFlagSet("list-payloads", flag.ExitOnError)
	payloadType := listPayloadsCmd.String("type", "schema", "Payload type to list")

	applyTransformCmd := flag.NewFlagSet("apply-transform", flag.ExitOnError)
	transformSpec := applyTransformCmd.String("spec", "", "Transform specification (e.g., 234:6/2:9)")

	// Add the database flag to each subcommand
	subcommands := []*flag.FlagSet{connectCmd, addSchemaCmd, addTransformerCmd, summaryCmd, listPayloadsCmd, applyTransformCmd}
	//for _, cmd := range subcommands {
	//	cmd.StringVar(&databaseFile, "database", ":memory:", "SQLite database file")
	//}

	if len(os.Args) < 2 {
		fmt.Println("Expected 'connect', 'add-schema', 'add-transformer', 'summary', 'list-payloads', or 'apply-transform' subcommands")
		for _, cmd := range subcommands {
			fmt.Printf("\nSubcommand '%s':\n", cmd.Name())
			cmd.PrintDefaults()
		}
		os.Exit(1)
	}

	fmt.Println("Database file:", databaseFile)
	restOfArgs := os.Args[parseArgsIndex:]
	if len(restOfArgs) < 1 {
		fmt.Println("Expected subcommand")
		os.Exit(1)
	}
	fmt.Println("rest of the args:", restOfArgs)

	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := initializeDatabase(db); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	switch os.Args[parseArgsIndex] {
	case "add-schema":
		addSchemaCmd.Parse(restOfArgs[1:])
		content := *schemaContent
		if strings.HasPrefix(content, "@") {
			fileContent, err := ioutil.ReadFile(strings.TrimPrefix(content, "@"))
			if err != nil {
				log.Fatal(err)
			}
			content = string(fileContent)
		}
		if err := addSchemaPayload(db, content); err != nil {
			log.Fatal(err)
		}
	case "add-transformer":
		addTransformerCmd.Parse(restOfArgs[1:])
		content := *transformerContent
		if strings.HasPrefix(content, "@") {
			fileContent, err := ioutil.ReadFile(strings.TrimPrefix(content, "@"))
			if err != nil {
				log.Fatal(err)
			}
			content = string(fileContent)
		}
		if err := addTransformerPayload(db, content); err != nil {
			log.Fatal(err)
		}
	case "summary":
		summaryCmd.Parse(restOfArgs[1:])
		stats, err := getSummaryStats(db)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Summary stats:")
		fmt.Printf("%v\n", stats)
		for payloadType, count := range stats {
			fmt.Printf("%s: %d\n", payloadType, count)
		}
	case "list-payloads":
		listPayloadsCmd.Parse(restOfArgs[1:])
		if err := listPayloads(db, *payloadType); err != nil {
			log.Fatal(err)
		}
	case "apply-transform":
		applyTransformCmd.Parse(restOfArgs[1:])
		parts := strings.Split(*transformSpec, "/")
		if len(parts) != 2 {
			log.Fatal("Invalid transform specification")
		}
		inputParts := strings.Split(parts[0], ":")
		outputParts := strings.Split(parts[1], ":")
		if len(inputParts) != 2 || len(outputParts) != 2 {
			log.Fatal("Invalid transform specification")
		}
		inputID, _ := strconv.ParseInt(inputParts[0], 10, 64)
		inputSchemaID, _ := strconv.ParseInt(inputParts[1], 10, 64)
		transformerID, _ := strconv.ParseInt(outputParts[0], 10, 64)
		outputSchemaID, _ := strconv.ParseInt(outputParts[1], 10, 64)
		if err := applyTransform(db, inputID, inputSchemaID, transformerID, outputSchemaID); err != nil {
			log.Fatal(err)
		}
	case "connect":
		connectCmd.Parse(restOfArgs[1:])
		opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", *mqttHost, *mqttPort))
		fmt.Println("Connecting to MQTT broker", *mqttHost, *mqttPort)
		opts.SetClientID(*clientName)

		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Printf("Error connecting to MQTT broker: %v", token.Error())
			log.Fatal(token.Error())
		}

		if token := client.Subscribe("#", 0, func(client mqtt.Client, msg mqtt.Message) {
			cpd := &CommonPayloadData{
				Time:        time.Now().Unix(),
				TimeAdded:   time.Now().Unix(),
				Device:      "mqtt-dumper",
				Topic:       msg.Topic(),
				PayloadType: "raw",
				Payload:     string(msg.Payload()),
			}

			if err := insertCommonPayloadData(db, cpd); err != nil {
				log.Printf("Error inserting data: %v", err)
			} else {
				log.Printf("Inserted: %+v", cpd)
			}
		}); token.Wait() && token.Error() != nil {
			log.Printf("Error subscribing to MQTT topic: %v", token.Error())
			log.Fatal(token.Error())
		}

		log.Println("Connected to MQTT broker. Press CTRL+C to exit.")
		select {}
	default:
		fmt.Println("Expected subcommand")
		// show available subcommands
		fmt.Println("Available subcommands:")
		fmt.Println("  add-schema - Add a schema payload")
		fmt.Println("  add-transformer - Add a transformer payload")
		fmt.Println("  summary - Show summary stats")
		fmt.Println("  list-payloads - List payloads")
		fmt.Println("  apply-transform - Apply a transform")
		fmt.Println("  connect - Connect to MQTT broker")
		connectCmd.Usage()
		os.Exit(1)
	}

}
