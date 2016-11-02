package metrics

import (
	"os"
	"fmt"
	"log"
	"time"
	"github.com/armon/go-metrics/datadog"
)

type StatsdClient struct {
	client *datadog.DogStatsdSink
}

func NewStatsdClient(udplistenerport int) (*StatsdClient, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}
	sc := new(StatsdClient)
	temp, clientErr := datadog.NewDogStatsdSink(fmt.Sprintf(":%d", udplistenerport), hostname)
	if clientErr != nil {
		return nil, err
	}
	sc.client = temp
	return sc, nil
}

func (s *StatsdClient) IncrementCounter(tags []string) {
	s.client.IncrCounterWithTags([]string{"usage","counter"}, 1, tags)
}

func (s *StatsdClient) SendElapsedTime(start time.Time, tags []string) {
	elapsedTimeInMilliSeconds := float32(time.Since(start).Nanoseconds()/(1000*1000));
	s.client.AddSampleWithTags([]string{"elapsed","time"}, elapsedTimeInMilliSeconds, tags)
}
