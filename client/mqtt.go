package client

import (
	"time"
	"errors"
	"strings"
	"crypto/tls"
	"k8s.io/klog"
	"github.com/jwzl/wssocket/model"
	"github.com/jwzl/wssocket/translator"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct{
	Host			string
	User, Passwd	string
	ClientID		string
	CleanSession 	bool
	keepAliveInterval time.Duration
	PingTimeout		  time.Duration	
	// optinal as below.
	OnConnect		mqtt.OnConnectHandler
	OnLost			mqtt.ConnectionLostHandler
	FileStorePath	string
	//Will message, optional
	WillTopic		string
	WillMessage		string
	WillQOS			byte
	WillRetained	bool	
	// tls config
	TLSConfig *tls.Config
	client 	mqtt.Client
}

func NewMQTTClient() *MQTTClient {
	return &MQTTClient{}
}

func (mc *MQTTClient) Start() {
	opts := mqtt.NewClientOptions()
	
	opts.AddBroker(mc.Host)
	opts.SetClientID(mc.ClientID)

	opts.SetUsername(mc.User)
	opts.SetPassword(mc.Passwd)

	opts.SetCleanSession(mc.CleanSession)
	if mc.TLSConfig != nil {
		klog.Infof("SSL/TLS is enabled!")
		opts.SetTLSConfig(mc.TLSConfig)
	}
	if strings.Compare(mc.FileStorePath, "memory") != 0 {
		klog.Infof("use file store!")
		opts.SetStore(mqtt.NewFileStore(mc.FileStorePath))
	}
	if mc.OnConnect != nil {
		opts.SetOnConnectHandler(mc.OnConnect)
	}
	if mc.OnLost != nil {
		opts.SetConnectionLostHandler(mc.OnLost)
	}
	if strings.Compare(mc.WillTopic, "") != 0{
		opts.SetWill(mc.WillTopic, mc.WillMessage, mc.WillQOS, mc.WillRetained)
	}
	opts.SetKeepAlive(mc.keepAliveInterval)
	opts.SetPingTimeout(mc.PingTimeout)

	// Create mqtt client.
	client := mqtt.NewClient(opts)
	mc.client = client
}

func (mc *MQTTClient) Connect() error {
	if mc.client == nil {
		return errors.New("nil client")
	}

	if token := mc.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

//Publish
// We Encode this model message and publishe it.
func (mc *MQTTClient) Publish(topic string, qos byte, retained bool, msg *model.Message) error {
	if msg == nil {
		return errors.New("msg is nil")	
	}
	
	if mc.client == nil {
		return errors.New("nil client")
	}

	if !mc.client.IsConnectionOpen() {
		return errors.New("connection is not active")	
	}  
	
	rawData, err := translator.NewTransCoding().Encode(msg)
	if err != nil {
		return err
	}

	if token := mc.client.Publish(topic, qos, retained, rawData); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

//Subscribe
// We Overide the function for the fn can process the model message.
func (mc *MQTTClient) Subscribe(topic string, qos byte, fn func(msg *model.Message)) error {
	if mc.client == nil {
		return errors.New("nil client")
	}

	if !mc.client.IsConnectionOpen() {
		return errors.New("connection is not active")	
	}  

	callback := func (client mqtt.Client, message mqtt.Message) {
		msg := &model.Message{}
		rawData := message.Payload()
	
		err := translator.NewTransCoding().Decode(rawData, msg)	
		if err != nil {
			klog.Infof("error message format, Ignored!")
			return 
		}

		fn(msg)
	}
	if token := mc.client.Subscribe(topic, qos, callback); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (mc *MQTTClient) Unsubscribe(topics string) error {
	if mc.client == nil {
		return errors.New("nil client")
	}

	if !mc.client.IsConnectionOpen() {
		return errors.New("connection is not active")	
	}  

	if token := mc.client.Unsubscribe(topics); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (mc *MQTTClient) Close() {
	mc.client.Disconnect(250) 
}
