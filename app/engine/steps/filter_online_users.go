package steps

import (
	botApp "neurobot/app/bot"
	"neurobot/model/payload"

	"github.com/apex/log"
)

type filterOnlineUsersRunner struct {
	eid         string
	botRegistry botApp.Registry
}

func (runner *filterOnlineUsersRunner) Run(p *payload.Payload) error {
	log.Log.WithFields(log.Fields{
		"executionID":  runner.eid,
		"workflowStep": "filterOnlineUsers",
	}).Info("running workflow step")

	mc, err := runner.botRegistry.GetPrimaryClient()
	if err != nil {
		return err
	}

	// loop through all users and get their presence status
	var onlineUsers []string
	for _, u := range p.Users {
		if mc.GetPresence(u) == "online" {
			onlineUsers = append(onlineUsers, u)
		}
	}

	p.Users = onlineUsers // effectively removing non-online users (offline or unknown)

	return nil
}

func NewFilterOnlineUsers(eid string, botRegistry botApp.Registry) *filterOnlineUsersRunner {
	return &filterOnlineUsersRunner{
		eid:         eid,
		botRegistry: botRegistry,
	}
}
