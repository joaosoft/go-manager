package gomanager

// INSQProducer ...
type INSQProducer interface {
	Start() error
	Stop() error
	Publish(topic string, body []byte, maxRetries int) error
	Ping() error
	Started() bool
}

// AddNSQProducer ...
func (manager *GoManager) AddNSQProducer(key string, nsqProducer INSQProducer) error {
	manager.nsqProducers[key] = nsqProducer
	log.Infof("nsq producer %s added", key)

	return nil
}

// RemoveNSQProducer ...
func (manager *GoManager) RemoveNSQProducer(key string) (INSQProducer, error) {
	process := manager.nsqProducers[key]

	delete(manager.processes, key)
	log.Infof("nsq producer %s removed", key)

	return process, nil
}

// GetNSQProducer ...
func (manager *GoManager) GetNSQProducer(key string) INSQProducer {
	if process, exists := manager.nsqProducers[key]; exists {
		return process
	}
	log.Infof("nsq producer %s doesn't exist", key)
	return nil
}
