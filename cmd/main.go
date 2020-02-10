package main

import (
	"fmt"
)

var messageChan chan string = make(chan string)
/*
var (
	eventSource string
	eventType   string
	sink        string
	messageChan chan string = make(chan string)
)

type envConfig struct {
	// Sink URL where to send IOT cloudevents
	Sink string `envconfig:"K_SINK"`

	// CEOverrides are the CloudEvents overrides to be applied to the outbound event.
	//CEOverrides string `envconfig:"K_CE_OVERRIDES"`

	// Name of this pod.
	Name string `envconfig:"POD_NAME" required:"true"`

	// Namespace this pod exists in.
	Namespace string `envconfig:"POD_NAMESPACE" required:"true"`

	// Whether to run continuously or exit.
	OneShot bool `envconfig:"ONE_SHOT" default:"false"`
}
*/

func main() { 

	//Load in Env variables 
	/*
	flag.Parse()

	var env envConfig
	if err := envconfig.Process("", &env); err != nill {
			log.Printf("[ERROR] Failed to process env var: %s",err)
			os.Exit(1)
	}

	if env.Sink != ""{
		sink = eng.Sink 
	}else{
		log.Println("No Sink Set")
	}

	c, err := kncloudevents.NewDefaultClient(sink)
	if err != nil {
		log.Fatalf("failed to create client: %s", err.Error())
	}

	if eventSource == "" {
		eventSource = fmt.Sprintf("https://knative.dev/eventing-contrib/cmd/iotContanerSource/#%s/%s", env.Namespace, env.Name)
		log.Printf("Heartbeats Source: %s", eventSource)
	}
	*/


	go consume("telemetry","amqps://messaging-8lxzny44dx-enmasse-infra.apps.astoycos-ocp.shiftstack.com:443","myapp.iot","consumer","foobar","tls.crt",2)

	for{
	
	msg := <- messageChan

	fmt.Println(msg)	

	}
}