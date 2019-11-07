package client

import (
	"k8s.io/klog"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct{
	Host			string
	
	ClientID		string
	CleanSession 	bool
	keepAliveInterval int64
	PingTimeout		  int64	
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
	
}

func (mc *MQTTClient) Start() {
	opts := mqtt.NewClientOptions()
	
	opts.AddBroker(mc.Host)
	opts.SetClientID(mc.ClientID)

	opts.SetUsername(user)
	opts.SetPassword(password)

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
