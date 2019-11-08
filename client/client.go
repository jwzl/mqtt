package client

import (
	"k8s.io/klog"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	MQTT_CLIENT_DISCONNECTED	= "Disconneted"
	MQTT_CLIENT_CONNECTED	= "Conneted"
)

type Client struct {
	Host			string
	ClientID		string
	keepAliveInterval int64
	PingTimeout		  int64	

	State 			string	
	// tls config
	tlsConfig 		*tls.Config
	client  		*MQTTClient
}

func NewClient() *Client{
	client := &Client{
		State: MQTT_CLIENT_DISCONNECTED,
	}

	// Create mqtt client.
	c := &MQTTClient{
		Host:			client.Host,
		ClientID:		client.ClientID,
		keepAliveInterval:	client.keepAliveInterval,
		PingTimeout:		client.PingTimeout,			
		CleanSession:	true,
		FileStorePath: "memory",
		OnConnect:	client.ClientOnConnect,
		OnLost:		client.ClientOnLost,			
		WillTopic:		"",			//no will topic.	
		TLSConfig:		client.tlsConfig,  
	}

	client.client = c

	return client 
}

func (c *Client) Start(){
	//Start the mqtt client
	c.client.Start()

	if err := c.client.Connect(); err != nil {
		klog.Infof("Client connecte err (%v)", err)
		return
	}

	klog.Infof("Client connecte Successful")
}

func (c *Client) Publish(topic string, msg *model.Message) error {
	return c.client.Publish(topic, 2, false, msg)
}

func (c *Client) Subscribe(topic string, fn func(msg *model.Message)) error {
	return c.client.Subscribe(topic, 2, fn)
}

func (c *Client) Unsubscribe(topics string) error {
	return c.client.Unsubscribe(topics)
}

func (c *Client) Close(){
	c.client.Close()
}

func (c *Client) ClientOnConnect(client mqtt.Client){
	klog.Infof("Client (%s) connected", c.ClientID)

	c.State = MQTT_CLIENT_CONNECTED
}

func (c *Client) ClientOnLost(client mqtt.Client, err error){
	klog.Infof("Client (%s) disconnected with error (%v)", c.ClientID, err)

	c.State = MQTT_CLIENT_DISCONNECTED
}
