package manager

import (
	"encoding/json"
	"fmt"
	"github.com/joaosoft/web"
	"github.com/joaosoft/logger"
	"github.com/labstack/gommon/log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"sync"

	"github.com/labstack/echo"
	"github.com/nsqio/go-nsq"
)

// --------- dummy process ---------
func dummy_process() error {
	logger.Info("hello, i'm exetuting the dummy process")
	return nil
}

// --------- dummy nsq ---------
type dummy_nsq_handler struct{}

func (dummy *dummy_nsq_handler) HandleMessage(msg *nsq.Message) error {
	logger.Infof("executing the handle message of NSQ with [ message: %s ]", string(msg.Body))
	return nil
}

// --------- dummy web http ---------
func dummy_web_http_handler(w http.ResponseWriter, r *http.Request) {
	type Example struct {
		Id   string `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	example := Example{Id: "123", Name: "joao", Age: 29}
	jsonIndent, _ := json.MarshalIndent(example, "", "    ")
	w.Write(jsonIndent)
}

// --------- dummy web echo ---------
func dummy_web_echo_handler(ctx echo.Context) error {
	type Example struct {
		Id   string `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	return ctx.JSON(http.StatusOK, Example{Id: ctx.Param("id"), Name: "joao", Age: 29})
}

func work_handler(id string, data interface{}) error {
	logger.Infof("work with the id %s and Data %s done!", id, data.(string))
	return nil
}

func usage() {
	//
	// Manager
	manager := NewManager()

	//
	// Manager: processes
	process := manager.NewSimpleProcess(dummy_process)
	if err := manager.AddProcess("process_1", process); err != nil {
		logger.Errorf("MAIN: error on processes %s", err)
	}

	//
	// nsq producer
	nsqConfigProducer := NewNSQConfig("topic_1", "channel_1", []string{"127.0.0.1:4150"}, []string{"127.0.0.1:4161"}, 30, 5)
	nsqProducer, _ := manager.NewSimpleNSQProducer(nsqConfigProducer)
	manager.AddNSQProducer("nsq_producer_1", nsqProducer)
	nsqProducer = manager.GetNSQProducer("nsq_producer_1")
	nsqProducer.Publish("topic_1", []byte("MENSAGEM ENVIADA PARA A NSQ"), 3)

	logger.Info("waiting 1 seconds...")
	<-time.After(time.Duration(1) * time.Second)

	//
	// Manager: nsq consumer
	nsqConfigConsumer := NewNSQConfig("topic_1", "channel_1", []string{"127.0.0.1:4161"}, []string{"127.0.0.1:4150"}, 30, 5)
	nsqConsumer, _ := manager.NewSimpleNSQConsumer(nsqConfigConsumer, &dummy_nsq_handler{})
	manager.AddProcess("nsq_consumer_1", nsqConsumer)

	//
	// Manager: configuration
	type dummy_config struct {
		App  string `json:"app"`
		User struct {
			Name   string `json:"name"`
			Age    int    `json:"age"`
			Random int    `json:"random"`
		} `json:"user"`
	}
	dir, _ := os.Getwd()
	obj := &dummy_config{}
	simpleConfig, _ := NewSimpleConfig(dir+"/bin/launcher/Data/config.json", obj)
	manager.AddConfig("config_1", simpleConfig)
	config := manager.GetConfig("config_1")

	jsonIndent, _ := json.MarshalIndent(config.GetObj(), "", "    ")
	logger.Infof("CONFIGURATION: %s", jsonIndent)

	// allows to set a new configuration and save in the file
	n := rand.Intn(9000)
	obj.User.Random = n
	logger.Infof("MAIN: Random: %d", n)
	config.Set(obj)
	if err := config.Save(); err != nil {
		logger.Error("MAIN: error whe saving configuration file")
	}

	//
	// Manager: simpleWeb

	// simpleWeb - with http
	simpleWeb := manager.NewSimpleWebHttp(":8081")
	if err := manager.AddWeb("web_http", simpleWeb); err != nil {
		logger.Error("error adding simpleWeb process to Manager")
	}
	simpleWeb = manager.GetWeb("web_http")
	simpleWeb.AddRoute(http.MethodGet, "/web_http", dummy_web_http_handler)

	// simpleWeb - with echo
	simpleWeb = manager.NewSimpleWebEcho(":8082")
	if err := manager.AddWeb("web_echo", simpleWeb); err != nil {
		logger.Error("error adding simpleWeb process to Manager")
	}
	simpleWeb = manager.GetWeb("web_echo")
	simpleWeb.AddRoute(http.MethodGet, "/web_echo/:id", dummy_web_echo_handler)
	go simpleWeb.Start(&sync.WaitGroup{}) // starting this because of the gateway

	logger.Info("waiting 1 seconds...")
	<-time.After(time.Duration(1) * time.Second)

	//
	// Manager: gateway
	headers := map[string][]string{"Content-Type": {"application/json"}}

	gateway, err := manager.NewSimpleGateway()
	if err != nil {
		log.Errorf("%s", err)
	}

	manager.AddGateway("gateway_1", gateway)
	gateway = manager.GetGateway("gateway_1")
	status, bytes, err := gateway.Request(http.MethodGet, "http://127.0.0.1:8082", "/web_echo/123", string(web.ContentTypeApplicationJSON), headers, nil)
	logger.Infof("status: %d, response: %s, error? %t", status, string(bytes), err != nil)

	//
	// Manager: database

	// database - postgres
	postgresConfig := NewDBConfig("postgres", "postgres://user:password@localhost:7001?sslmode=disable")
	postgresConn := manager.NewSimpleDB(postgresConfig)
	manager.AddDB("postgres", postgresConn)

	// database - mysql
	mysqlConfig := NewDBConfig("mysql", "root:password@tcp(127.0.0.1:7002)/mysql")
	mysqlConn := manager.NewSimpleDB(mysqlConfig)
	manager.AddDB("mysql", mysqlConn)

	//
	// Manager: redis
	redisConfig := NewRedisConfig("127.0.0.1", 7100, 0, "")
	redisConn := manager.NewSimpleRedis(redisConfig)
	manager.AddRedis("redis", redisConn)

	//
	// Manager: workqueue
	workqueueConfig := NewWorkListConfig("queue_001", 1, 2, time.Second*2, FIFO)
	workqueue := manager.NewSimpleWorkList(workqueueConfig, work_handler, nil, nil)
	manager.AddWorkList("queue_001", workqueue)
	workqueue = manager.GetWorkList("queue_001")
	for i := 1; i <= 1000; i++ {
		go workqueue.AddWork(fmt.Sprintf("PROCESS: %d", i), fmt.Sprintf("THIS IS MY MESSAGE %d", i))
	}
	if err := workqueue.Start(&sync.WaitGroup{}); err != nil {
		logger.Errorf("MAIN: error on workqueue %s", err)
	}

	manager.Start()
}
