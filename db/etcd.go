/*
   ToDD databasePackage implementation for etcd

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package db

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"

	"github.com/Mierdin/todd/agent/defs"
	"github.com/Mierdin/todd/config"
	"github.com/Mierdin/todd/server/objects"
)

// newEtcdDB is a factory function that produces a new instance of etcdDB with the configuration
// loaded and ready to be used.
func newEtcdDB(cfg config.Config) *etcdDB {
	var edb etcdDB
	edb.config = cfg
	return &edb
}

type etcdDB struct {
	config config.Config
}

func (etcddb etcdDB) Init() {

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	resp, err := kapi.Get(context.Background(), "/todd/agents", &client.GetOptions{Recursive: true})

	if resp != nil {

		log.Info("Deleting '/todd/agents' key")
		resp, err = kapi.Delete(context.Background(), "/todd/agents", &client.DeleteOptions{Recursive: true, Dir: true})

		if err != nil {
			log.Fatal(err)
		} else {
			// print common key info
			log.Printf("Set is done. Metadata is %q\n", resp)
		}
	}

	// TODO(mierdin): Consider deleting the entire /todd key here, and recreating all of the various subkeys, just to start from scratch.
	// Shouldn't need any previous data if you restart the server.
}

// SetAgent will ingest an agent advertisement, and update or insert the agent record
// in the database as needed.
func (etcddb etcdDB) SetAgent(adv defs.AgentAdvert) {

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	log.Infof("Setting '/todd/agents/%s' key", adv.Uuid)

	adv_json, err := json.Marshal(adv)
	if err != nil {
		log.Error("Problem converting Agent Advertisement to JSON")
		os.Exit(1)
	}

	// TODO(mierdin): TTL needs to be user-configurable
	resp, err := kapi.Set(
		context.Background(),                      // context
		fmt.Sprintf("/todd/agents/%s", adv.Uuid),  // key
		string(adv_json),                          // value
		&client.SetOptions{TTL: time.Second * 30}, //optional args
	)
	if err != nil {
		log.Error("Problem setting agent in etcd")
		os.Exit(1)
	}

	log.Infof("Agent set in etcd. This advertisement is good until %s", resp.Node.Expiration)

}

// GetAgent will retrieve a specific agent from the database by UUID
func (etcddb etcdDB) GetAgent(uuid string) defs.AgentAdvert {

	adv := defs.AgentAdvert{}

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	log.Printf("Getting /todd/agents/%s' key value", uuid)

	key_str := fmt.Sprintf("/todd/agents/%s", uuid)

	resp, err := kapi.Get(context.Background(), key_str, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Errorf("Agent %s not found.", uuid)
		return adv
	}

	log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)

	// Marshal API data into object
	err = json.Unmarshal([]byte(resp.Node.Value), &adv)
	if err != nil {
		log.Error("Failed to unmarshal json into agent advertisement")
		os.Exit(1)
	}

	// We want to use the TTLDuration field from etcd for simplicity
	adv.Expires = resp.Node.TTLDuration()

	// The etcd key should always match the inner JSON, so let's panic (for now) if this ever happens
	if uuid != adv.Uuid {
		panic("UUID in etcd does not match inner JSON text")
	}

	return adv

}

// GetAgents will retrieve all agents from the database
func (etcddb etcdDB) GetAgents() []defs.AgentAdvert {

	var ret_adv []defs.AgentAdvert

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	log.Print("Getting /todd/agents' key value")

	key_str := "/todd/agents"

	resp, err := kapi.Get(context.Background(), key_str, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Warn("Agent list empty when queried")
	} else {
		log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)

		if !resp.Node.Dir {
			panic("Expected dir in etcd for agents - encountered single node")
		}

		for i := range resp.Node.Nodes {
			node := resp.Node.Nodes[i]

			// Extract UUID from key string
			uuid := strings.Replace(node.Key, "/todd/agents/", "", 1)

			// Marshal API data into object
			var adv defs.AgentAdvert
			err = json.Unmarshal([]byte(node.Value), &adv)
			if err != nil {
				log.Error("Failed to retrieve data from server")
				os.Exit(1)
			}

			// We want to use the TTLDuration field from etcd for simplicity
			adv.Expires = node.TTLDuration()

			// The etcd key should always match the inner JSON, so let's panic (for now) if this ever happens
			if uuid != adv.Uuid {
				panic("UUID in etcd does not match inner JSON text")
			}

			ret_adv = append(ret_adv, adv)

		}

	}

	return ret_adv

}

// RemoveAgent will delete an agent advertisement present in etcd. This function exists for the rare situation when
// an Agent needs to be removed immediately, as opposed to simply waiting for the TTL to expire.
func (etcddb etcdDB) RemoveAgent(adv defs.AgentAdvert) {

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	resp, err := kapi.Get(context.Background(), fmt.Sprintf("/todd/agents/%s", adv.Uuid), &client.GetOptions{Recursive: true})

	if resp != nil {

		resp, err = kapi.Delete(context.Background(), fmt.Sprintf("/todd/agents/%s", adv.Uuid), &client.DeleteOptions{Recursive: true, Dir: true})

		if err != nil {
			log.Fatal(err)
		} else {
			log.Infof("Removed '/todd/agents/%s' key", adv.Uuid)
		}
	}
}

// GetObjects retrieves a list of ToddObjects stored within etcd, and returns this as a slice.
// This requires an "obj_type" string to specify the type of object being looked up.
func (etcddb etcdDB) GetObjects(obj_type string) []objects.ToddObject {

	var ret_obj []objects.ToddObject

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	// Construct the path to the key depending on the obj_type param
	var key_str string
	if obj_type == "" {
		// TODO(mierdin): support this empty type, and return an entire list of objects regardless of type
		log.Warn("Object API queried with no type argument -- returning empty slice")
		var empty []objects.ToddObject
		return empty
	} else {
		//Construct a path to the key based on the provided type
		key_str = fmt.Sprintf("/todd/objects/%s/", obj_type)
	}

	log.Info("Accessing objects at", key_str)

	resp, err := kapi.Get(context.Background(), key_str, &client.GetOptions{Recursive: true})
	if err != nil {
		fmt.Println(err)
		log.Warn("ToDD object store empty when queried")
	} else {

		log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)

		// We are expecting that this node is a directory
		if resp.Node.Dir {

			// Iterate over found objects
			for i := range resp.Node.Nodes {
				node := resp.Node.Nodes[i]

				log.Printf("Parsing object %s \n", node.Value)

				// Marshal API data into ToddObject
				var baseobj objects.BaseObject
				err = json.Unmarshal([]byte(node.Value), &baseobj)
				if err != nil {
					log.Error("Failed to retrieve data from server")
					os.Exit(1)
				}

				// Generate a more specific Todd Object based on the JSON data
				finalobj := baseobj.ParseToddObject([]byte(node.Value))

				ret_obj = append(ret_obj, finalobj)

			}
		} else {

			// We are definitely expecting an ETCd directory, so we should return nothing if this is not the case.
			log.Warn("Etcd query for objects did not result in a directory as expected -- returning empty slice")
			var empty []objects.ToddObject
			return empty

		}
	}

	return ret_obj
}

// SetObject will insert or update a ToddObject within etcd
func (etcddb etcdDB) SetObject(tobj objects.ToddObject) {

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	obj_json, err := json.Marshal(tobj)
	if err != nil {
		log.Error("Problem converting object to JSON")
		os.Exit(1)
	}

	log.Debugf("Setting '/todd/objects/%s/%s' key", tobj.GetType(), tobj.GetLabel())

	// Here, we set the key string, using the following format:
	// /todd/objects/<type>/<label(name)>
	key_str := fmt.Sprintf("/todd/objects/%s/%s", tobj.GetType(), tobj.GetLabel())

	_, err = kapi.Set(
		context.Background(), // context
		key_str,              // key
		string(obj_json),     // value
		nil,                  //optional args
	)
	if err != nil {
		log.Error("Problem setting object in etcd")
		os.Exit(1)
	}

	log.Infof("Wrote new Todd Object to etcd: %s/%s", tobj.GetType(), tobj.GetLabel())

}

// DeleteObject will delete a ToddObject from etcd
func (etcddb etcdDB) DeleteObject(label string, objtype string) {

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	resp, err := kapi.Get(context.Background(), fmt.Sprintf("/todd/objects/%s/%s", objtype, label), &client.GetOptions{Recursive: true})

	if resp != nil {

		resp, err = kapi.Delete(context.Background(), fmt.Sprintf("/todd/objects/%s/%s", objtype, label), &client.DeleteOptions{Recursive: true, Dir: true})

		if err != nil {
			log.Fatal(err)
		} else {
			log.Infof("Removed '/todd/objects/%s/%s' key", objtype, label)
		}
	}
}

// SetGroupMapping will update etcd with the results of a grouping calculation
func (etcddb etcdDB) SetGroupMap(groupmap map[string]string) {

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	gmap_json, err := json.Marshal(groupmap)
	if err != nil {
		log.Error("Problem converting group map to JSON")
		os.Exit(1)
	}

	log.Debug("Setting '/todd/groupmap' key")

	key_str := "/todd/groupmap"

	_, err = kapi.Set(
		context.Background(), // context
		key_str,              // key
		string(gmap_json),    // value
		nil,                  //optional args
	)
	if err != nil {
		log.Error("Problem setting group map in etcd")
		os.Exit(1)
	}

	log.Infof("Updated group map in etcd: %s", gmap_json)

}

// GetGroupMap returns a map containing agent-to-group mappings. Agent UUIDs are used for keys
func (etcddb etcdDB) GetGroupMap() map[string]string {

	var ret_map map[string]string

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	key_str := "/todd/groupmap"

	log.Debug("Retrieving group map")

	resp, err := kapi.Get(context.Background(), key_str, &client.GetOptions{Recursive: true})
	if err != nil {
		fmt.Println(err)
		log.Warn("Error retrieving group mapping")
		return map[string]string{}
	}

	// Marshal etcd data into map
	err = json.Unmarshal([]byte(resp.Node.Value), &ret_map)
	if err != nil {
		log.Error("Failed to retrieve group map from etcd")
		os.Exit(1)
	}

	return ret_map
}

// InitTestRun is responsible for initializing a new test run within the database. This includes creating an entry for the test itself
// using the provided UUID for uniqueness, but also in the case of etcd, a nested entry for each agent participating in the test. Each
// Agent entry will be initially populated with that agent's current group and an initial status, but it will also house the result of
// that agent's testrun data, which will be aggregate dafter all agents have checked back in.
func (etcddb etcdDB) InitTestRun(testUuid string, testAgentMap map[string]map[string]string) error {
	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	// Create high-level UUID key for this testrun
	log.Debug("Creating entry in etcd for testrun ", testUuid)
	_, err = kapi.Set(
		context.Background(),                       // context
		fmt.Sprintf("/todd/testruns/%s", testUuid), // key
		"", // value
		&client.SetOptions{Dir: true, TTL: time.Second * 3000}, //optional args
		// TODO(mierdin): I set the TTL here so that I didn't dirty etcd with a bunch of old testruns while I develop this feature.
		// Need to decide if doing our own garbage collection is a better approach.
	)
	if err != nil {
		log.Error("Problem setting testrun UUID: ", testUuid)
		os.Exit(1)
	}

	// Create agent entry for each agent that is in the provided map
	for _, uuidmappings := range testAgentMap {

		// _ is either "targets" or "sources".
		// uuidmappings is a map[string]string that contains uuid (key) to group name (value) mappings for this test.

		for agent, group := range uuidmappings {
			// Create agent entry within this testrun
			log.Debugf("Creating agent entry within testrun %s for agent %s", testUuid, agent)
			_, err = kapi.Set(
				context.Background(),                                        // context
				fmt.Sprintf("/todd/testruns/%s/agents/%s", testUuid, agent), // key
				"", // value
				&client.SetOptions{Dir: true}, //optional args
			)
			if err != nil {
				log.Error("Problem setting initial agent placeholder in testrun: ", testUuid)
				log.Error(err)
				os.Exit(1)
			}

			var initAgentProps = map[string]string{
				"group":  group,
				"status": "init",
				// Intentially omitting the "testdata" key here, because we will create it when the testdata is ready
			}

			for k, v := range initAgentProps {

				_, err = kapi.Set(
					context.Background(),                                              // context
					fmt.Sprintf("/todd/testruns/%s/agents/%s/%s", testUuid, agent, k), // key
					v,   // value
					nil, //optional args
				)
				if err != nil {
					log.Error("Problem setting initial agent placeholder in testrun: ", testUuid)
					log.Error(err)
					os.Exit(1)
				}
			}
		}

	}

	return nil
}

// SetAgentTestStatus sets the status for an agent in a particular testrun key.
func (etcddb etcdDB) SetAgentTestStatus(testUuid, agentUuid, status string) error {
	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	_, err = kapi.Set(
		context.Background(),                                                   // context
		fmt.Sprintf("/todd/testruns/%s/agents/%s/status", testUuid, agentUuid), // key
		status, // value
		nil,    //optional args
	)
	if err != nil {
		log.Errorf("Problem updating status for agent %s in test %s", agentUuid, testUuid)
		log.Error(err)
		os.Exit(1)
	}

	return nil
}

// SetAgentTestData sets the post-test data for an agent in a particular testrun
func (etcddb etcdDB) SetAgentTestData(testUuid, agentUuid, testData string) error {
	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	_, err = kapi.Set(
		context.Background(),                                                     // context
		fmt.Sprintf("/todd/testruns/%s/agents/%s/testdata", testUuid, agentUuid), // key
		testData, // value
		nil,      //optional args
	)
	if err != nil {
		log.Errorf("Problem updating testdata for agent %s in test %s", agentUuid, testUuid)
		log.Error(err)
		os.Exit(1)
	}

	return nil
}

// GetTestStatus returns a map containing a list of agent UUIDs that are participating in the provided test, and their status in this test.
func (etcddb etcdDB) GetTestStatus(testUuid string) map[string]string {

	ret_map := make(map[string]string)

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	key_str := fmt.Sprintf("/todd/testruns/%s/agents", testUuid)

	log.Debug("Retrieving detailed test status for ", testUuid)

	resp, err := kapi.Get(context.Background(), key_str, &client.GetOptions{Recursive: true})
	if err != nil {
		fmt.Println(err)
		log.Errorf("Error - empty test encountered: %s", testUuid)
		return ret_map
	} else {

		log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)

		// We are expecting that this node is a directory
		if resp.Node.Dir {

			// Iterate over found objects
			for i := range resp.Node.Nodes {
				thisAgent := resp.Node.Nodes[i]

				statusKey := fmt.Sprintf("%s/status", thisAgent.Key)

				// Extract UUID from key string
				agentUuid := strings.Replace(thisAgent.Key, fmt.Sprintf("/todd/testruns/%s/agents/", testUuid), "", 1)

				statusResp, err := kapi.Get(context.Background(), statusKey, nil)
				if err != nil {
					log.Errorf("Error - empty agent status encountered: %s", testUuid)
				}

				ret_map[agentUuid] = statusResp.Node.Value

			}

		} else {
			log.Warn("Etcd query for detailed test status did not result in a directory as expected -- returning empty map")
			empty := make(map[string]string)
			return empty

		}
	}

	return ret_map
}

// GetAgentTestData returns un-sanitized data from the individual agents. For a report of all agents' data,
// which has been sanitized by the server, see GetCleanTestData
func (etcddb etcdDB) GetAgentTestData(testUuid, sourceGroup string) map[string]string {

	ret_map := make(map[string]string)

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	key_str := fmt.Sprintf("/todd/testruns/%s/agents", testUuid)

	log.Debug("Retrieving detailed test data for ", testUuid)

	resp, err := kapi.Get(context.Background(), key_str, &client.GetOptions{Recursive: true})
	if err != nil {
		fmt.Println(err)
		log.Errorf("Error - empty test encountered: %s", testUuid)
		return ret_map
	} else {

		log.Debugf("Etcd 'get' is done. Metadata is %q\n", resp)

		// We are expecting that this node is a directory
		if resp.Node.Dir {

			// Iterate over found objects
			for i := range resp.Node.Nodes {
				thisAgent := resp.Node.Nodes[i]

				// Extract UUID from key string
				agentUuid := strings.Replace(thisAgent.Key, fmt.Sprintf("/todd/testruns/%s/agents/", testUuid), "", 1)

				groupKey := fmt.Sprintf("%s/group", thisAgent.Key)
				groupResp, err := kapi.Get(context.Background(), groupKey, nil)
				if err != nil {
					log.Errorf("Error retrieving group of agent in: %s", testUuid)
				}

				if groupResp.Node.Value == sourceGroup {
					testRunDataKey := fmt.Sprintf("%s/testdata", thisAgent.Key)
					dataResp, err := kapi.Get(context.Background(), testRunDataKey, nil)
					if err != nil {
						log.Errorf("Error retrieving testdata of agent in: %s", testUuid)
					}
					ret_map[agentUuid] = dataResp.Node.Value
				}
			}
		} else {
			log.Warn("Etcd query for detailed test data did not result in a directory as expected -- returning empty map")
			empty := make(map[string]string)
			return empty

		}
	}
	return ret_map
}

// WriteCleanTestData will write the post-test metrics data that has been cleaned up and
// ready to be displayed or exported to the database
func (etcddb etcdDB) WriteCleanTestData(testUuid string, testData string) {

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	log.Debugf("/todd/testruns/%s/cleandata/", testUuid)

	key_str := fmt.Sprintf("/todd/testruns/%s/cleandata/", testUuid)

	_, err = kapi.Set(
		context.Background(), // context
		key_str,              // key
		testData,             // value
		nil,                  //optional args
	)
	if err != nil {
		log.Error("Problem setting object in etcd")
		os.Exit(1)
	}

	log.Infof("Wrote clean test data to test uuid: %s", testUuid)

}

// GetCleanTestData will retrieve clean test data from the database
func (etcddb etcdDB) GetCleanTestData(testUuid string) string {

	etcd_loc := fmt.Sprintf("http://%s:%s", etcddb.config.DB.IP, etcddb.config.DB.Port)
	cfg := client.Config{
		Endpoints: []string{etcd_loc},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	key_str := fmt.Sprintf("/todd/testruns/%s/cleandata", testUuid)

	log.Debug("Retrieving clean test data for ", testUuid)

	resp, err := kapi.Get(context.Background(), key_str, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Error(err)
		log.Errorf("Error - empty test data: %s", testUuid)
		return ""
	}

	return string(resp.Node.Value)
}

// TODO (mierdin): I have commented this out for now - may use this in the future to ensure that only one test is activated at a time.
//
// SetFlag will update etcd with the flag that indicates if tests can be run.
// func (etcddb EtcdDB) SetFlag(s string) errors.Error {

//     // Enforce the two valid values for this function
//     if s != "nogo" && s != "go" {
//         return errors.New("Invalid value for SetFlag")
//     }

//     cfg := client.Config{
//         Endpoints: []string{"http://192.168.59.103:4001"},
//         Transport: client.DefaultTransport,
//         // set timeout per request to fail fast when the target endpoint is unavailable
//         HeaderTimeoutPerRequest: time.Second,
//     }
//     c, err := client.New(cfg)
//     if err != nil {
//         log.Fatal(err)
//     }
//     kapi := client.NewKeysAPI(c)

//     _, err = kapi.Set(
//         context.Background(), // context
//         "/todd/flag",         // key
//         string(s),            // value
//         nil,                  //optional args
//     )
//     common.FailOnError(err, "Problem setting flag in etcd")

//     log.Infof("Set '/todd/flag' key to %s", s)

//     return nil
// }
