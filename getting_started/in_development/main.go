package in_development

import (
	"github.com/labstack/gommon/log"
	queue "go-manager/queue"
)

func main() {
	log.Infof("JOB START")

	shutdownChannelIn := make(chan bool)
	workChannelBufferSize := 5
	repository := queue.Repository{}
	queueController := queue.NewQueueController(repository)
	myqueue := queue.NewQueue(shutdownChannelIn, workChannelBufferSize, queueController)

	bytes := []byte(`a, b, c`)
	work := queue.NewWork(bytes)
	myqueue.AddWork(work)

	bytes = []byte(`d, e, f`)
	work = queue.NewWork(bytes)
	myqueue.AddWork(work)

	<-shutdownChannelIn

	log.Infof("JOB END")
}