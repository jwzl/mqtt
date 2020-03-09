package client

import (
	"time"
	"errors"
	"crypto/tls"
	"k8s.io/klog"
	"github.com/jwzl/wssocket/model"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	MQTT_CLIENT_DISCONNECTED	= "Disconneted"
	MQTT_CLIENT_CONNECTED	= "Conneted"
	retryCount       = 5
	cloudAccessSleep = 5 * time.Second
)

type Client struct {
	// scheme://host:port
	// Where "scheme" is one of "tcp", "ssl", or "ws", "host" is the ip-address (or hostname)
	// and "port" is the port on which the broker is accepting connections.
	Host			string
	User, Passwd	string
	// the client id to be used by this client when
	// connecting to the MQTT broker.
	ClientID		string
	// the amount of time (in seconds) that the client
	// should wait before sending a PING request to the broker.
	// default as 120s.
	keepAliveInterval time.Duration
	// the amount of time (in seconds) that the client
	// will wait after sending a PING request to the broker.
	// default as 120s 
	PingTimeout		  time.Duration
	//the size of the internal queue that holds messages while the
    // client is temporairily offline.
	MessageCacheDepth  uint
	//0: QOSAtMostOnce, 1: QOSAtLeastOnce, 2: QOSExactlyOnce.
	QOS				   byte
	//if the flag set true, server will store the message and 
	//can be delivered to future subscribers.
	Retain			   bool	
	//the state of client.
	State 			string	
	// tls config
	tlsConfig 		*tls.Config
	client  		*MQTTClient
}

func NewClient(host, user, passwd, clientID string) *Client{
	if host == "" || clientID == "" {
		return nil
	}

	client := &Client{
		Host:  host,
		User: user,
		Passwd: passwd,
		ClientID: clientID,
		keepAliveInterval: 120 * time.Second,
		PingTimeout: 120 * time.Second,		
		State: MQTT_CLIENT_DISCONNECTED,
		MessageCacheDepth : 100,
		QOS:		byte(2),
		Retain:		false,	
	}

	return client 
}

func (c *Client) SetkeepAliveInterval(k time.Duration) {
	c.keepAliveInterval = k
}

func (c *Client) SetPingTimeout(k time.Duration) {
	c.PingTimeout = k
}

func (c *Client) SetTlsConfig (config *tls.Config) {
	c.tlsConfig = config
}
func (c *Client) SetQOS (qos byte) {
	c.QOS = qos
}
func (c *Client) SetMessageCacheDepth(depth uint) {
	c.MessageCacheDepth = depth
}
func (c *Client) SetRetain (ret bool) {
	c.Retain = ret
}

func (c *Client) Start() error {
	// Create mqtt client.
	client := &MQTTClient{
		Host:			c.Host,
		User:			c.User,
		Passwd:			c.Passwd,	
		ClientID:		c.ClientID,
		Order: 			false,	
		keepAliveInterval:	c.keepAliveInterval,
		PingTimeout:		c.PingTimeout,
		MessageChannelDepth: c.MessageCacheDepth,			
		CleanSession:	true,
		FileStorePath: "memory",
		OnConnect:	c.ClientOnConnect,
		OnLost:		c.ClientOnLost,			
		WillTopic:		"",			//no will topic.	
		TLSConfig:		c.tlsConfig,  
	}

	c.client = client

	//Start the mqtt client
	c.client.Start()

	for i := 0; i < retryCount; i++ {
		err := c.client.Connect()
		if err != nil {
			klog.Errorf("Client connecte err (%v), retry...", err)
		}else {
			klog.Infof("Client connecte Successful")
			return nil
		}
		time.Sleep(cloudAccessSleep)
	}

	klog.Errorf("max retry count reached when connecting to cloud")
	return errors.New("max retry count reached when connecting to cloud")
}

func CreateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	return tlsConfig, nil
}

func (c *Client) Publish(topic string, msg *model.Message) error {
	return c.client.Publish(topic, c.QOS, c.Retain, msg)
}

func (c *Client) Subscribe(topic string, fn func(topic string, msg *model.Message)) error {
	return c.client.Subscribe(topic, c.QOS, fn)
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
