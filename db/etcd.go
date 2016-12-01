/*
   ToDD databasePackage implementation for etcd

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package db

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/server/objects"
)

const (
	keyAgents   = "/todd/agents"
	keyGroupMap = "/todd/groupmap"
	keyObjects  = "/todd/objects"
	keyTestRuns = "/todd/testruns"
)

func init() {
	register("etcd", newEtcdDB)
}

// newEtcdDB is a factory function that produces a new instance of etcdDB with the configuration
// loaded and ready to be used.
func newEtcdDB(cfg *config.Config) (Database, error) {
	db := &etcdDB{config: cfg}

	if err := db.init(); err != nil {
		return nil, err
	}

	return db, nil
}

type etcdDB struct {
	config  *config.Config
	keysAPI client.KeysAPI
}

func (db *etcdDB) init() error {
	url := fmt.Sprintf("http://%s:%s", db.config.DB.Host, db.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{url},
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		return errors.Wrap(err, "creating new etcd client")
	}

	db.keysAPI = client.NewKeysAPI(c)

	// TODO(mierdin): Consider deleting the entire /todd key here, and recreating all of the various subkeys, just to start from scratch.
	// Shouldn't need any previous data if you restart the server.
	err = db.remove(keyAgents)
	if err != nil && err != ErrNotExist {
		return err
	}

	return nil
}

// SetAgent will ingest an agent advertisement, and update or insert the agent record
// in the database as needed.
func (db *etcdDB) SetAgent(adv defs.AgentAdvert) error {
	key := path.Join(keyAgents, adv.UUID)
	resp, err := db.setJSON(key, adv, &client.SetOptions{TTL: time.Second * 30}) // TODO(mierdin): TTL needs to be user-configurable
	if err == nil {
		log.Infof("Agent set in etcd. This advertisement is good until %s", resp.Node.Expiration)
	}
	return err
}

// GetAgent will retrieve a specific agent from the database by UUID
func (db *etcdDB) GetAgent(uuid string) (*defs.AgentAdvert, error) {
	key := path.Join(keyAgents, uuid)
	node, err := db.get(key)
	if err != nil {
		if err == ErrNotExist {
			log.Errorf("Agent %s not found.", uuid)
		}
		return nil, err
	}

	return nodeToAgentAdvert(node, uuid)
}

// nodeToAgentAdvert takes a etcd Node representing an AgentAdvert and returns an AgentAdvert
func nodeToAgentAdvert(node *client.Node, expectedUUID string) (*defs.AgentAdvert, error) {
	// Marshal API data into object
	adv := defs.AgentAdvert{
		Expires: node.TTLDuration(), // We want to use the TTLDuration field from etcd for simplicity
	}
	err := json.Unmarshal([]byte(node.Value), &adv)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling to JSON")
	}

	// The etcd key should always match the inner JSON
	if expectedUUID != adv.UUID {
		return nil, errors.New("UUID in etcd does not match inner JSON text")
	}

	return &adv, nil
}

// GetAgents will retrieve all agents from the database
func (db *etcdDB) GetAgents() ([]defs.AgentAdvert, error) {
	adverts := []defs.AgentAdvert{}

	node, err := db.get(keyAgents)
	if err != nil {
		if err == ErrNotExist {
			log.Warn("Agent list empty when queried")
			return adverts, nil
		}
		return nil, err
	}

	if !node.Dir {
		return nil, errors.New("Expected dir in etcd for agents - encountered single node")
	}

	for _, node := range node.Nodes {
		// Extract UUID from key string
		uuid := path.Base(node.Key)

		adv, err := nodeToAgentAdvert(node, uuid)
		if err != nil {
			return nil, err
		}

		adverts = append(adverts, *adv)
	}

	return adverts, nil

}

// RemoveAgent will delete an agent advertisement present in etcd. This function exists for the rare situation when
// an Agent needs to be removed immediately, as opposed to simply waiting for the TTL to expire.
func (db *etcdDB) RemoveAgent(uuid string) error {
	return db.remove(path.Join(keyAgents, uuid))
}

// GetObjects retrieves a list of ToddObjects stored within etcd, and returns this as a slice.
// This requires an "objType" string to specify the type of object being looked up.
func (db *etcdDB) GetObjects(objType string) ([]objects.ToddObject, error) {
	retObj := []objects.ToddObject{}

	// Construct the path to the key depending on the objType param
	if objType == "" {
		// TODO(mierdin): support this empty type, and return an entire list of objects regardless of type
		return nil, errors.New("Object API queried with no type argument")
	}

	//Construct a path to the key based on the provided type
	node, err := db.get(path.Join(keyObjects, objType))
	if err != nil {
		if err == ErrNotExist {
			log.Warn("ToDD object store empty when queried")
			return retObj, nil
		}
		return nil, err
	}

	// We are expecting that this node is a directory
	if !node.Dir {
		// We are definitely expecting an ETCd directory, so we should return nothing if this is not the case.
		return nil, errors.New("Etcd query for objects did not result in a directory as expected")
	}

	// Iterate over found objects
	for _, node := range node.Nodes {
		log.Printf("Parsing object %s", node.Value)

		// Marshal API data into ToddObject
		var baseobj objects.BaseObject
		err = json.Unmarshal([]byte(node.Value), &baseobj)
		if err != nil {
			return nil, err
		}

		// Generate a more specific Todd Object based on the JSON data
		finalobj, err := baseobj.ParseToddObject([]byte(node.Value))
		if err != nil {
			return nil, err
		}

		retObj = append(retObj, finalobj)
	}

	return retObj, nil
}

// SetObject will insert or update a ToddObject within etcd
func (db *etcdDB) SetObject(tobj objects.ToddObject) error {
	key := path.Join(keyObjects, tobj.GetType(), tobj.GetLabel())
	_, err := db.setJSON(key, tobj, nil)
	if err != nil {
		log.Error("Problem setting object in etcd")
	}
	return err
}

// DeleteObject will delete a ToddObject from etcd
func (db *etcdDB) DeleteObject(label string, objType string) error {
	return db.remove(path.Join(keyObjects, objType, label))
}

// SetGroupMapping will update etcd with the results of a grouping calculation
func (db *etcdDB) SetGroupMap(groupmap map[string]string) error {
	_, err := db.setJSON(keyGroupMap, groupmap, nil)
	if err != nil {
		log.Error("Problem setting group map in etcd")
	}
	return err
}

// GetGroupMap returns a map containing agent-to-group mappings. Agent UUIDs are used for keys
func (db *etcdDB) GetGroupMap() (map[string]string, error) {
	retMap := map[string]string{}

	node, err := db.get(keyGroupMap)
	if err != nil {
		log.Warnf("Error retrieving group mapping: %v", err)
		return retMap, nil
	}

	// Marshal etcd data into map
	err = json.Unmarshal([]byte(node.Value), &retMap)
	return retMap, errors.Wrap(err, "unmarshaling JSON")
}

// InitTestRun is responsible for initializing a new test run within the database. This includes creating an entry for the test itself
// using the provided UUID for uniqueness, but also in the case of etcd, a nested entry for each agent participating in the test. Each
// Agent entry will be initially populated with that agent's current group and an initial status, but it will also house the result of
// that agent's testrun data, which will be aggregate dafter all agents have checked back in.
func (db *etcdDB) InitTestRun(testUUID string, testAgentMap map[string]map[string]string) error {
	// Create high-level UUID key for this testrun
	log.Debug("Creating entry in etcd for testrun ", testUUID)
	testrunKey := path.Join(keyTestRuns, testUUID)
	err := db.createDir(testrunKey)
	if err != nil {
		return errors.Wrapf(err, "creating directory %q", testrunKey)
	}

	// TODO(vcabbage): Parallelize this?
	// Create agent entry for each agent that is in the provided map
	for _, uuidmappings := range testAgentMap {
		// _ is either "targets" or "sources".
		// uuidmappings is a map[string]string that contains uuid (key) to group name (value) mappings for this test.
		for agent, group := range uuidmappings {
			// Create agent entry within this testrun
			log.Debugf("Creating agent entry within testrun %s for agent %s", testUUID, agent)
			key := path.Join(testrunKey, "agents", agent)
			err = db.createDir(key)
			if err != nil {
				return errors.Wrapf(err, "creating directory %q", key)
			}

			initAgentProps := map[string]string{
				"group":  group,
				"status": "init",
				// Intentially omitting the "testdata" key here, because we will create it when the testdata is ready
			}

			for k, v := range initAgentProps {
				propKey := path.Join(key, k)
				err = db.set(propKey, v)
				if err != nil {
					return errors.Wrapf(err, "setting %q = %q", propKey, v)
				}
			}
		}
	}

	return nil
}

// SetAgentTestStatus sets the status for an agent in a particular testrun key.
func (db *etcdDB) SetAgentTestStatus(testUUID, agentUUID, status string) error {
	key := path.Join(keyTestRuns, testUUID, "agents", agentUUID, "status")
	return db.set(key, status)
}

// SetAgentTestData sets the post-test data for an agent in a particular testrun
func (db *etcdDB) SetAgentTestData(testUUID, agentUUID string, testData map[string]map[string]interface{}) error {
	key := path.Join(keyTestRuns, testUUID, "agents", agentUUID, "testdata")
	_, err := db.setJSON(key, testData, nil)
	return err
}

// GetTestStatus returns a map containing a list of agent UUIDs that are participating in the provided test, and their status in this test.
func (db *etcdDB) GetTestStatus(testUUID string) (map[string]string, error) {
	retMap := make(map[string]string)

	key := path.Join(keyTestRuns, testUUID, "agents")

	node, err := db.get(key)
	if err != nil {
		log.Errorf("Error - empty test encountered for %q: %v", testUUID, err)
		return retMap, nil
	}

	// We are expecting that this node is a directory
	if !node.Dir {
		return nil, errors.New("Etcd query for detailed test status did not result in a directory as expected")
	}

	// Iterate over found objects
	for _, node := range node.Nodes {
		agentUUID := path.Base(node.Key) // Extract UUID from key string
		for _, subNode := range node.Nodes {
			if path.Base(subNode.Key) == "status" {
				retMap[agentUUID] = subNode.Value
			}
		}
	}

	return retMap, nil
}

// GetAgentTestData returns un-sanitized data from the individual agents. For a report of all agents' data,
// which has been sanitized by the server, see GetCleanTestData
func (db *etcdDB) GetTestData(testUUID string) (map[string]map[string]map[string]interface{}, error) {
	retMap := make(map[string]map[string]map[string]interface{})

	node, err := db.get(path.Join(keyTestRuns, testUUID, "agents"))
	if err != nil {
		return nil, err
	}

	// We are expecting that this node is a directory
	if !node.Dir {
		return nil, errors.New("Etcd query for detailed test data did not result in a directory as expected")
	}

	// Iterate over found objects
	for _, node := range node.Nodes {
		// Extract UUID from key string
		agentUUID := path.Base(node.Key)

		for _, subNode := range node.Nodes {
			if path.Base(subNode.Key) != "testdata" {
				continue
			}

			var testData map[string]map[string]interface{}
			err := json.Unmarshal([]byte(subNode.Value), &testData)
			if err != nil {
				log.Errorf("Unable to unmarshal testData for %q: %s", agentUUID, subNode.Value)
				continue
			}
			retMap[agentUUID] = testData
		}
	}

	return retMap, nil
}

func (db *etcdDB) createDir(key string) error {
	log.Infof("Creating directory %q in etcd", key)

	_, err := db.keysAPI.Set(
		context.Background(), // context
		key,                  // key
		"",                   // value
		&client.SetOptions{Dir: true, TTL: time.Second * 3000}, //optional args
		// TODO(mierdin): I set the TTL here so that I didn't dirty etcd with a bunch of old testruns while I develop this feature.
		// Need to decide if doing our own garbage collection is a better approach.
	)
	return err
}

func (db *etcdDB) get(key string) (*client.Node, error) {
	log.Printf("Getting %q key value", key)

	resp, err := db.keysAPI.Get(context.Background(), key, &client.GetOptions{Recursive: true})
	if err != nil {
		if isNotExist(err) {
			return nil, ErrNotExist
		}
		return nil, err
	}

	log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)
	return resp.Node, nil
}

func (db *etcdDB) remove(key string) error {
	_, err := db.keysAPI.Delete(context.Background(), key, &client.DeleteOptions{Recursive: true, Dir: true})
	if err != nil {
		if isNotExist(err) {
			return ErrNotExist
		}
		return err
	}

	log.Infof("Removed %q key", key)
	return nil
}

func (db *etcdDB) set(key, value string) error {
	log.Debugf("Setting %q = %q", key, value)

	_, err := db.keysAPI.Set(
		context.Background(), // context
		key,                  // key
		value,                // value
		nil,                  //optional args
	)
	return err
}

func (db *etcdDB) setJSON(key string, v interface{}, opts *client.SetOptions) (*client.Response, error) {
	log.Infof("Setting %q key", key)

	j, err := json.Marshal(v)
	if err != nil {
		log.Error("Problem converting Agent Advertisement to JSON")
		return nil, errors.Wrap(err, "marshaling to JSON")
	}

	resp, err := db.keysAPI.Set(
		context.Background(), // context
		key,                  // key
		string(j),            // value
		opts,                 //optional args
	)

	return resp, errors.Wrap(err, "setting in etcd")
}

func isNotExist(err error) bool {
	etcdErr, ok := err.(client.Error)
	if !ok {
		return false
	}
	return etcdErr.Code == client.ErrorCodeKeyNotFound
}
