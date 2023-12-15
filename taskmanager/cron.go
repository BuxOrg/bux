package taskmanager

import "github.com/robfig/cron/v3"

// cronLocal is the interface for the "local cron" service
type cronLocal struct {
	cronService *cron.Cron
}

// New will stop any existing cron service and start a new one
func (c *cronLocal) New() {
	if c.cronService != nil {
		c.cronService.Stop()
	}
	c.cronService = cron.New()
}

// AddFunc will add a function to the cron service
func (c *cronLocal) AddFunc(spec string, cmd func()) (int, error) {
	e, err := c.cronService.AddFunc(spec, cmd)
	return int(e), err
}

// Start will start the cron service
func (c *cronLocal) Start() {
	c.cronService.Start()
}

// Stop will stop the cron service
func (c *cronLocal) Stop() {
	c.cronService.Stop()
}
