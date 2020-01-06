package main

import (
	"time"
	"k8s.io/klog"
	"github.com/jwzl/mqtt/client"
	"github.com/jwzl/wssocket/model"
)

func main(){

	client := client.NewClient("tcp://172.21.73.155:1883", "", "", "edgeNode1")
	//Start the mqtt client.
	client.Start()

	client.Subscribe("mqtt/dgtwin/cloud/#", func(topic string, msg *model.Message){

		klog.Infof("message is arrived %v", msg)
	})

	for {
		msg := &model.Message{}
		now := time.Now().UnixNano() / 1e6
		msg.BuildRouter("device", "", "twin", "twin", "Read")
		msg.BuildHeader("12345", now)
		client.Publish("mqtt/dgtwin/cloud/me", msg)

		time.Sleep(2 *time.Second)
	}
}
