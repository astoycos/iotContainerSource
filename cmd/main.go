package main

import (
	"fmt"
	"context"
	"flag"
	"log"
	"os"
	"knative.dev/eventing-contrib/pkg/kncloudevents"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
)

type dataPayload struct {
	Temp string    `json:"temp"`
	Scale    string `json:"scale"`
}

var (
	eventSource string
	eventType   string
	sink        string
	messageChan chan string = make(chan string) //# of devices Connected to System 
)

func init() {
	flag.StringVar(&eventSource, "eventSource", "", "the event-source (CloudEvents)")
	flag.StringVar(&eventType, "eventType", "github.com/astoycos/iotContainerSource", "the event-type (CloudEvents)")
	flag.StringVar(&sink, "sink", "", "the service url to send data to to")
}

type envConfig struct {
	// Sink URL where to send IOT cloudevents
	Sink string `envconfig:"K_SINK"`

	// Name of this pod.
	Name string `envconfig:"POD_NAME" required:"true"`

	// Namespace this pod exists in.
	Namespace string `envconfig:"POD_NAMESPACE" required:"true"`

	// Whether to run continuously or exit.
	OneShot bool `envconfig:"ONE_SHOT" default:"false"`

	//Enmasse Endpoint configs

	//Messaging endpoing URI  
	MessageURI string `envconfig:"MESSAGE_URI" required:"true"`

	//Message Type telemetry/event 
	MessageType string `envconfig:"MESSAGE_TYPE" required:"true"`

	//Messaging Tenant 
	MessageTenant string `envconfig:"MESSAGE_TENANT" required:"true"`

	//TLS Setting 0:No TLS 1: TLS INSECURE 2:TLS Secure 
	TLSConfig int `envconfig:"TLS_CONFIG" default:"0"`

	//Path to TLS credentials 
	TLSPath string `envconfig:"TLS_PATH" default:""`

	//hono client Username
	ClientUsername string `envconfig:"CLIENT_USERNAME" default:""`
	
	//hono client Password 
	ClientPassword string `envconfig:"CLIENT_PASSWORD" default:"" `
}


func main() { 

	//Load in Env variables 
	
	flag.Parse()

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
			log.Printf("[ERROR] Failed to process env var: %s",err)
			os.Exit(1)
	}
	
	fmt.Println(env.MessageURI)
	fmt.Println(env.Name)

	if env.Sink != ""{
		sink = env.Sink 
	}else{
		log.Println("Sink not set")
	}

	c, err := kncloudevents.NewDefaultClient(sink)
	if err != nil {
		log.Fatalf("failed to create client: %s", err.Error())
	}

	if eventSource == "" {
		eventSource = fmt.Sprintf("https://github.com/astoycos/iotContanerSource/#%s/%s", env.Namespace, env.Name)
		log.Printf("IOT data Source: %s", eventSource)
	}

	dp := &dataPayload{
		Temp: "",
		Scale: "Farenheight",
	}

	go consume(env.MessageType,env.MessageURI,env.MessageTenant,env.ClientUsername,env.ClientPassword,env.TLSConfig,env.TLSPath)

	//Infinite loop that waits for data from edge and then forwards it to knative service 
	for{
		fmt.Println("Waiting for iot Device data...")
		//Channel to funnel messages from consuming goroutine 
		msg := <- messageChan	
		
		fmt.Println("Device Data received: " + msg)
		dp.Temp = msg; 
		fmt.Println("Making cloud Event")
		event := cloudevents.NewEvent("1.0")
		event.SetType(eventType)
		event.SetSource(eventSource)
		event.SetDataContentType(cloudevents.ApplicationJSON)

		if err := event.SetData(dp); err != nil {
			log.Printf("failed to set cloudevents data: %s", err.Error())
		}

		log.Printf("sending cloudevent to %s", sink)
		if _, _, err := c.Send(context.Background(), event); err != nil {
			log.Printf("failed to send cloudevent: %s", err.Error())
		}
		
		//ENV option to only receive once message
		if env.OneShot {
			return
		}	

	}
}