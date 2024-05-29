package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
	UserName   string
	Password   string
	ClientID   string
	BrokerURL  string
	BrokerPort string
	CertPath   string
	Topics     map[string]byte
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	var cfg Config

	flag.StringVar(&cfg.BrokerURL, "U", "localhost", "MQTT Broker URL")
	flag.StringVar(&cfg.BrokerPort, "P", "1883", "MQTT Broker Port")
	flag.StringVar(&cfg.UserName, "u", "", "MQTT Username")
	flag.StringVar(&cfg.Password, "p", "", "MQTT Password")
	flag.StringVar(&cfg.ClientID, "cid", "", "ClientID")
	flag.StringVar(&cfg.CertPath, "cert", "", "Path to your certificate file")
	flag.Parse()

	cfg.Topics = map[string]byte{ // replace with your topics
		"topic/device/temperature": 0,
		"topic/device/speed":       0,
		"topic/device/pressure":    0,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s:%s", cfg.BrokerURL, cfg.BrokerPort))
	opts.SetClientID(cfg.ClientID)
	opts.SetUsername(cfg.UserName)
	opts.SetPassword(cfg.Password)

	tlsConfig := newTLSConfig(cfg.CertPath)
	opts.SetTLSConfig(tlsConfig)

	opts.OnConnect = func(c mqtt.Client) {
		log.Println("subscriber connected")
		subscribe(c, cfg)
	}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("subscriber lost connection: %v\n", err)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("subscriber failed to connect: %v", token.Error())
	}

	<-c
	client.Disconnect(2000) // gracefully close connection after 2 seconds
}

func subscribe(client mqtt.Client, subClient Config) {
	token := client.SubscribeMultiple(subClient.Topics, func(c mqtt.Client, m mqtt.Message) {
		log.Printf("Received message: %s from topic %s\n", m.Payload(), m.Topic())
	})
	token.Wait()
	log.Printf("subscribed to topics: %v", subClient.Topics)
}

func newTLSConfig(certPath string) *tls.Config {
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(certPath) // replace with the path to your certificate
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	certPool.AppendCertsFromPEM(ca)
	return &tls.Config{
		RootCAs: certPool,
	}
}
