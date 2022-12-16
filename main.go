package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CivilTime time.Time
type Time time.Time

var db *gorm.DB
var err error

func (c *CivilTime) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`) //get rid of "
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse("2006-01-02T15:04:05.000000", value) //parse time
	if err != nil {
		return err
	}
	*c = CivilTime(t) //set result using the pointer
	// *c = t
	// *c = Time(t)
	return nil
}

// func (c CivilTime) String() string {
// 	// return fmt.Sprintf("%v (%v years)", c.Name, c.Age)
// 	var t time.Time = (time.Time)c
// 	return t.Format("2006")
// }

// func (PhoneGeo) TableName() string {
// 	return "phone_geo"
// }

type PhoneGeo struct {
	gorm.Model
	// Id        uint32         `json:"id,omitempty" `
	DeviceId  string  `json:"device_id" `
	Name      string  `json:"device_name" `
	Latitude  float64 `json:"latitude" `
	Longitude float64 `json:"longitude" `
	Speed     float32 `json:"speed" `
	Timestamp uint64  `json:"timestamp" `
	// EventTs   CivilTime `json:"event_ts" `
	// EventTs time.Time `json:"event_ts" `
	// EventTs string `json:"timestamp"`
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

	var geo PhoneGeo
	_ = json.Unmarshal([]byte(msg.Payload()), &geo)
	fmt.Println("Got geo unmarshalled")
	fmt.Println(geo.DeviceId)

	// fmt.Printf("timestamp %v\n", geo.Timestamp)
	// fmt.Printf("EventTs %v\n", geo.EventTs)

	// var t time.Time = (time.Time)(geo.EventTs)
	// fmt.Printf("geo.EventTs: %v\n", t.Year()

	result := db.Create(&geo) // pass pointer of data to Create
	fmt.Printf("gorm create result: %v\n", result)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	fmt.Println("Main")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mqttServer := os.Getenv("MQTT_SERVER")
	mqttUser := os.Getenv("MQTT_USER")
	mqttPassword := os.Getenv("MQTT_PASSWORD")
	mqttTopic := os.Getenv("MQTT_TOPIC")
	dbConn := os.Getenv("DB_CONN")

	fmt.Printf("%s %s %s %s\n", mqttServer, mqttUser, mqttPassword, mqttTopic)
	fmt.Printf("%s\n", dbConn)

	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", mqttServer, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(mqttUser)
	opts.SetPassword(mqttPassword)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	db, err = gorm.Open(mysql.Open(dbConn), &gorm.Config{})
	db.AutoMigrate(&PhoneGeo{})
	if err != nil {
		panic("failed to connect database")
	} else if db != nil {
		fmt.Println("Got db but what's next?")
	}

	sub(client, mqttTopic)

	time.Sleep(10 * time.Second)
	client.Disconnect(250)
}

func sub(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic %s", topic)
}
