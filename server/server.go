package server

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/responses"
	"github.com/toddproject/todd/agent/tasks"
	"github.com/toddproject/todd/comms"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/db"
	"github.com/toddproject/todd/hostresources"
)

type Server struct {
	config *config.Config
	comms  comms.Comms
	db     db.DatabasePackage
	assets assetProvider
}

func New(cfg *config.Config, cms comms.Comms, d db.DatabasePackage, assets assetProvider) *Server {
	return &Server{
		config: cfg,
		comms:  cms,
		db:     d,
		assets: assets,
	}
}

type assetProvider interface {
	Assets() map[string]map[string]string
}

func (s *Server) HandleAgentAdvertisement(body []byte) error {
	log.Debugf("Agent advertisement received: %s", body)

	var agent defs.AgentAdvert
	err := json.Unmarshal(body, &agent)
	if err != nil {
		return errors.Wrap(err, "unmarshaling JSON")
	}

	// assetList is a slice that will contain any URLs that need to be sent to an
	// agent as a response to an incorrect or incomplete list of assets
	var assetList []string

	// assets is the asset map from the SERVER's perspective
	for assetType, assetHashes := range s.assets.Assets() {
		// agentAssets is the asset map from the AGENT's perspective
		var agentAssets map[string]string
		switch assetType {
		case "factcollectors":
			agentAssets = agent.FactCollectors
		case "testlets":
			agentAssets = agent.Testlets
		default:
			log.Warn("Invalid asset type:", assetType)
			continue
		}

		for name, hash := range assetHashes {
			// See if the hashes match (a missing asset will also result in False)
			if agentAssets[name] == hash {
				continue
			}

			// hashes do not match, so we need to append the asset download URL to the remediate list
			defaultIP := s.config.LocalResources.IPAddrOverride
			if defaultIP == "" {
				defaultIP = hostresources.GetIPOfInt(s.config.LocalResources.DefaultInterface).String()
			}
			assetURL := fmt.Sprintf("http://%s:%s/%s/%s", defaultIP, s.config.Assets.Port, assetType, name)
			assetList = append(assetList, assetURL)
		}
	}

	if len(assetList) != 0 {
		log.Warnf("Agent %s did not have the required asset files. This advertisement is ignored.", agent.UUID)
		task := &tasks.DownloadAsset{
			BaseTask: tasks.BaseTask{Type: "DownloadAsset"},
			Assets:   assetList,
		}
		s.comms.SendTask(agent.UUID, task)
	}

	// Asset list is empty, so we can continue
	return s.db.SetAgent(agent)
}

func (s *Server) HandleAgentResponse(body []byte) error {
	log.Debugf("Agent response received: %s", body)

	// Unmarshal into BaseResponse to determine type
	var baseMsg responses.Base
	err := json.Unmarshal(body, &baseMsg)
	if err != nil {
		return errors.Wrap(err, "unmarshaling JSON")
	}

	// call agent response method based on type
	switch baseMsg.Type {
	case responses.KeySetAgentStatus:

		var sasr responses.SetAgentStatus
		err = json.Unmarshal(body, &sasr)
		if err != nil {
			log.Error("Problem unmarshalling AgentStatus")
		}

		log.Debugf("Agent %s is '%s' regarding test %s. Writing to DB.", sasr.AgentUUID, sasr.Status, sasr.TestUUID)
		err := s.db.SetAgentTestStatus(sasr.TestUUID, sasr.AgentUUID, sasr.Status)
		if err != nil {
			log.Errorf("Error writing agent status to DB: %v", err)
		}

	case responses.KeyUploadTestData:

		var utdr responses.UploadTestData
		err = json.Unmarshal(body, &utdr)
		if err != nil {
			log.Error("Problem unmarshalling UploadTestDataResponse")
		}

		err = s.db.SetAgentTestData(utdr.TestUUID, utdr.AgentUUID, utdr.TestData)
		if err != nil {
			log.Error("Problem setting agent test data")
		}

		// Send task to the agent that says to delete the entry
		dtdt := &tasks.DeleteTestData{
			BaseTask: tasks.BaseTask{Type: "DeleteTestData"},
			TestUUID: utdr.TestUUID,
		}
		s.comms.SendTask(utdr.AgentUUID, dtdt)

		// Finally, set the status for this agent in the test to "finished"
		err := s.db.SetAgentTestStatus(dtdt.TestUUID, utdr.AgentUUID, "finished")
		if err != nil {
			log.Errorf("Error writing agent status to DB: %v", err)
		}

	default:
		return errors.Errorf("Unexpected type value for received response: %s", baseMsg.Type)
	}
	return nil
}
