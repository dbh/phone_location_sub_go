package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var dbErr error

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("MQTT Connected")
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("MQTT Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

	var geo PhoneGeo
	err := json.Unmarshal([]byte(msg.Payload()), &geo)
	if err != nil {
		fmt.Printf("messagePubHandler failure during unmarshal %v\n", err)
		return
	}
	fmt.Println("Got geo unmarshalled")
	result := db.Create(&geo)
	fmt.Printf("gorm create result: %v\n", result)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("MQTT Connect lost: %v\n", err)
}

func mqttSubscribe(client mqtt.Client, mqttTopic string) {
	if token := client.Subscribe(mqttTopic, 0, nil); token.Wait() &&
		token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	fmt.Printf("Subscribed to topic %s\n", mqttTopic)
}

func main() {
	fmt.Println("Main")
	c := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mqttServer := os.Getenv("MQTT_SERVER")
	mqttUser := os.Getenv("MQTT_USER")
	mqttPassword := os.Getenv("MQTT_PASSWORD")
	mqttTopic := os.Getenv("MQTT_TOPIC")
	dbConn := os.Getenv("DB_CONN")

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

	db, dbErr = gorm.Open(mysql.Open(dbConn), &gorm.Config{})
	db.AutoMigrate(&PhoneGeo{})
	if dbErr != nil || db == nil {
		panic("failed to connect database")
	}

	mqttSubscribe(client, mqttTopic)

	go func() {
		sig := <-c
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	fmt.Println("awaiting signal")
	<-done
	fmt.Println("exiting")
}
