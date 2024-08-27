package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"os"
	"strings"

	//"io/ioutil"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-sql-driver/mysql"
)

// Struct containing the message information that is inserted into the database. Primary key and last measured timestamp are updated automatically so they are not included
type Message struct {
	topic_sensor_name string
	// gasName           string
	// units             string
	measurement string
}

// type Sensor struct {
// 	sensorID   int64
// 	sensorName string
// 	address   string
// }

// type Topic struct {
// 	topicID int64
// 	// sensorID          int64
// 	topicName         string
// 	gasName           string
// 	unitOfMeasurement string
// }

// type Log struct {
// 	logID       int64
// 	topicID     int64
// 	measurement string
// }

var db *sql.DB

// Update constants for specific deployment being connected to
const protocol = "ssl"
const broker = "g332f11e.ala.eu-central-1.emqxsl.com"
const port = 8883
const topic = "root/faux/data/#"
const username = "Rashaad"
const password = "Rashaad"

func main() {
	client := createMqttClient()
	go subscribe(client)         // we use goroutine to run the subscription function
	time.Sleep(time.Second * 10) // pause minimum of 2 seconds to wait for the subscription function to be ready, otherwise subscriber function doesn't receive messages
	var broker_msg, broker_topic, sensor_name = subscribe(client)
	fmt.Println(sensor_name)
	cfg := mysql.Config{
		User:                 "rashaad", //os.Getenv("DBUSER"), //Set DBUSER and DBPASS environment variables
		Passwd:               "RA5",     //os.Getenv("DBPASS"), //Alternatively, the SQL username and password can also be set manually without using environment variables but that will make them visible to the public if published to a public repository
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "emqx_data",
		AllowNativePasswords: true,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	msg := Message{broker_topic, broker_msg}
	tableInsert(db, msg)

}

func createMqttClient() mqtt.Client {
	connectAddress := fmt.Sprintf("%s://%s:%d", protocol, broker, port)
	clientID := "emqx_cloude096fd"

	fmt.Println("connect address: ", connectAddress)
	opts := mqtt.NewClientOptions()

	opts.AddBroker(connectAddress)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetClientID(clientID)
	opts.SetKeepAlive(time.Second * 60)

	opts.SetTLSConfig(loadTLSConfig("emqxsl-ca.pem"))
	opts.SetTLSConfig(loadTLSConfig("main.go"))

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.WaitTimeout(10*time.Second) && token.Error() != nil {
		log.Printf("\nConnection error: %s\n", token.Error())
	}
	return client
}

func subscribe(client mqtt.Client) (string, string, string) {
	qos := 0
	broker_msg := make(chan string)
	broker_topic := make(chan string)
	sensor_name := make(chan string)
	client.Subscribe(topic, byte(qos), func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message: %s, from topic: %s \n", msg.Payload(), msg.Topic())
		split := strings.Split(string(msg.Topic()), "/")
		broker_msg <- string(msg.Payload())
		broker_topic <- string(msg.Topic())
		sensor_name <- split[3]
	})

	return (<-broker_msg), (<-broker_topic), (<-sensor_name)
}

func loadTLSConfig(caFile string) *tls.Config {
	// load tls config
	var tlsConfig tls.Config
	tlsConfig.InsecureSkipVerify = true
	if caFile != "" {
		certpool := x509.NewCertPool()
		ca, err := os.ReadFile(caFile)
		if err != nil {
			log.Fatal(err.Error())
		}
		certpool.AppendCertsFromPEM(ca)
		tlsConfig.RootCAs = certpool
	}
	return &tlsConfig
}

func tableInsert(db *sql.DB, message Message) int {
	topicIDQuery := `SELECT topicID FROM Topics WHERE topicName = ?`
	topicID, err := db.Query(topicIDQuery, message.topic_sensor_name) // change to QueryRow
	if err != nil {
		log.Fatal(err)
	}

	// insert topic into Topics (if needed) and log into Logs
	logInsertQuery := `INSERT INTO Logs (topicID, measurement, measureTime)`
	if topicID == nil {
		topicInsertQuery := `INSERT INTO Topics (topic_name)
		VALUES (?)`
		topic_entry, err := db.Exec(topicInsertQuery, message.topic_sensor_name)
		if err != nil {
			log.Fatal(err)
		}

		topicID, err := topic_entry.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		log_entry, err := db.Exec(logInsertQuery, topicID, message.measurement)
		if err != nil {
			log.Fatal(err)
		}

		//Gets the ID for the
		lastInsertId, err := log_entry.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		return int(lastInsertId)

	} else {
		log_entry, err := db.Exec(logInsertQuery, topicID, message.measurement)
		if err != nil {
			log.Fatal(err)
		}

		//Gets the ID for the
		lastInsertId, err := log_entry.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		return int(lastInsertId)
	}
}
