/*
   ToDD databasePackage implementation for RabbitMQ

   Copyright 2015 - Matt Oswalt
*/

package db

import (
    "encoding/json"
    "fmt"
    "strings"
    "time"

    log "github.com/mierdin/todd/Godeps/_workspace/src/github.com/Sirupsen/logrus"
    "github.com/mierdin/todd/Godeps/_workspace/src/github.com/coreos/etcd/client"
    "golang.org/x/net/context"

    "github.com/mierdin/todd/agent/defs"
    "github.com/mierdin/todd/common"
    "github.com/mierdin/todd/config"
)

type EtcdDB struct {
    config config.Config
}

func (etcddb EtcdDB) Init() {
    cfg := client.Config{
        Endpoints: []string{"http://192.168.59.103:4001"},
        Transport: client.DefaultTransport,
        // set timeout per request to fail fast when the target endpoint is unavailable
        HeaderTimeoutPerRequest: time.Second,
    }
    c, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    kapi := client.NewKeysAPI(c)

    resp, err := kapi.Get(context.Background(), "/agents", &client.GetOptions{Recursive: true})

    if resp != nil {

        log.Info("Deleting '/agents' key")
        resp, err = kapi.Delete(context.Background(), "/agents", &client.DeleteOptions{Recursive: true, Dir: true})

        if err != nil {
            log.Fatal(err)
        } else {
            // print common key info
            log.Printf("Set is done. Metadata is %q\n", resp)
        }
    }
}

// SetAgent will ingest an agent advertisement, and either update/insert the agent record
// in the database as needed.
func (etcddb EtcdDB) SetAgent(adv defs.AgentAdvert) {

    cfg := client.Config{
        Endpoints: []string{"http://192.168.59.103:4001"},
        Transport: client.DefaultTransport,
        // set timeout per request to fail fast when the target endpoint is unavailable
        HeaderTimeoutPerRequest: time.Second,
    }
    c, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    kapi := client.NewKeysAPI(c)

    log.Infof("Setting '/agents/%s' key", adv.Uuid)

    adv_json, err := json.Marshal(adv)
    common.FailOnError(err, "Problem converting Agent Advertisement to JSON")

    // TODO(mierdin): TTL needs to be user-configurable
    resp, err := kapi.Set(
        context.Background(),                      // context
        fmt.Sprintf("/agents/%s", adv.Uuid),       // key
        string(adv_json),                          // value
        &client.SetOptions{TTL: time.Second * 30}, //optional args
    )
    common.FailOnError(err, "Problem setting agent in etcd")

    log.Infof("Agent set in etcd. This advertisement is good until %s", resp.Node.Expiration)

}

// GetAgents will retrieve all agents from the database
func (etcddb EtcdDB) GetAgents(uuid string) []defs.AgentAdvert {

    var ret_adv []defs.AgentAdvert

    cfg := client.Config{
        Endpoints: []string{"http://192.168.59.103:4001"},
        Transport: client.DefaultTransport,
        // set timeout per request to fail fast when the target endpoint is unavailable
        HeaderTimeoutPerRequest: time.Second,
    }
    c, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    kapi := client.NewKeysAPI(c)

    log.Print("Getting '/agents' key value")

    // Construct the path to the key depending on the UUID param
    var key_str string
    if uuid != "" {
        key_str = fmt.Sprintf("/agents/%s", uuid)
    } else {
        key_str = "/agents"
    }
    resp, err := kapi.Get(context.Background(), key_str, &client.GetOptions{Recursive: true})
    if err != nil {
        log.Warn("Agent list empty when queried")
    } else {
        // print common key info
        log.Printf("Get is done. Metadata is %q\n", resp)
        // print value
        log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)

        if resp.Node.Dir {

            for i := range resp.Node.Nodes {
                node := resp.Node.Nodes[i]

                // Extract UUID from key string
                uuid := strings.Replace(node.Key, "/agents/", "", 1)

                // Marshal API data into object
                var adv defs.AgentAdvert
                err = json.Unmarshal([]byte(node.Value), &adv)
                common.FailOnError(err, "Failed to retrieve data from server")

                // We want to use the TTLDuration field from etcd for simplicity
                adv.Expires = node.TTLDuration()

                // The etcd key should always match the inner JSON, so let's panic (for now) if this ever happens
                if uuid != adv.Uuid {
                    panic("UUID in etcd does not match inner JSON text")
                }

                ret_adv = append(ret_adv, adv)

            }
        } else {

            // Extract UUID from key string
            uuid := strings.Replace(resp.Node.Key, "/agents/", "", 1)

            // Marshal API data into object
            var adv defs.AgentAdvert
            err = json.Unmarshal([]byte(resp.Node.Value), &adv)
            common.FailOnError(err, "Failed to retrieve data from server")

            // We want to use the TTLDuration field from etcd for simplicity
            adv.Expires = resp.Node.TTLDuration()

            // The etcd key should always match the inner JSON, so let's panic (for now) if this ever happens
            if uuid != adv.Uuid {
                panic("UUID in etcd does not match inner JSON text")
            }

            ret_adv = append(ret_adv, adv)

        }
    }

    return ret_adv

}

// RemoveAgent will delete an agent advertisement present in etcd
func (etcddb EtcdDB) RemoveAgent(adv defs.AgentAdvert) {
    cfg := client.Config{
        Endpoints: []string{"http://192.168.59.103:4001"},
        Transport: client.DefaultTransport,
        // set timeout per request to fail fast when the target endpoint is unavailable
        HeaderTimeoutPerRequest: time.Second,
    }
    c, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    kapi := client.NewKeysAPI(c)

    resp, err := kapi.Get(context.Background(), fmt.Sprintf("/agents/%s", adv.Uuid), &client.GetOptions{Recursive: true})

    if resp != nil {

        resp, err = kapi.Delete(context.Background(), fmt.Sprintf("/agents/%s", adv.Uuid), &client.DeleteOptions{Recursive: true, Dir: true})

        if err != nil {
            log.Fatal(err)
        } else {
            log.Infof("Removed '/agents/%s' key", adv.Uuid)
        }
    }
}

func (etcddb EtcdDB) SetObject(obj_type string, label string, content interface{}) {

    // TODO(mierdin): Have some kind of enum here that detects the type of obj

    cfg := client.Config{
        Endpoints: []string{"http://192.168.59.103:4001"},
        Transport: client.DefaultTransport,
        // set timeout per request to fail fast when the target endpoint is unavailable
        HeaderTimeoutPerRequest: time.Second,
    }
    c, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    kapi := client.NewKeysAPI(c)

    log.Debugf("Setting '/toddobjects/%s/%s' key", obj_type, label)

    adv_json, err := json.Marshal(matches)
    common.FailOnError(err, "Problem converting object to JSON")

    // TODO(mierdin): remove this TTL
    _, err = kapi.Set(
        context.Background(),                               // context
        fmt.Sprintf("/toddobjects/%s/%s", obj_type, label), // key
        string(adv_json),                                   // value
        &client.SetOptions{TTL: time.Second * 30},          //optional args
    )
    common.FailOnError(err, "Problem setting object in etcd")

    log.Infof("NEW GROUP FILE - %s/%s", obj_type, label)

}

func (etcddb EtcdDB) GetGroupFiles() {
    var ret_adv []defs.AgentAdvert

    cfg := client.Config{
        Endpoints: []string{"http://192.168.59.103:4001"},
        Transport: client.DefaultTransport,
        // set timeout per request to fail fast when the target endpoint is unavailable
        HeaderTimeoutPerRequest: time.Second,
    }
    c, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    kapi := client.NewKeysAPI(c)

    log.Print("Getting '/groupfiles' key value")

    resp, err := kapi.Get(context.Background(), "/groupfiles", &client.GetOptions{Recursive: true})
    if err != nil {
        log.Warn("Group file empty when queried")
    } else {
        // print common key info
        log.Printf("Get is done. Metadata is %q\n", resp)
        // print value
        log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)

        if resp.Node.Dir {

            for i := range resp.Node.Nodes {
                node := resp.Node.Nodes[i]

                // Extract UUID from key string
                uuid := strings.Replace(node.Key, "/groupfiles/", "", 1)

                // Marshal API data into object
                var adv defs.AgentAdvert
                err = json.Unmarshal([]byte(node.Value), &adv)
                common.FailOnError(err, "Failed to retrieve data from server")

                // We want to use the TTLDuration field from etcd for simplicity
                adv.Expires = node.TTLDuration()

                // The etcd key should always match the inner JSON, so let's panic (for now) if this ever happens
                if uuid != adv.Uuid {
                    panic("UUID in etcd does not match inner JSON text")
                }

                ret_adv = append(ret_adv, adv)

            }
        } else {

            // Extract UUID from key string
            uuid := strings.Replace(resp.Node.Key, "/groupfiles/", "", 1)

            // Marshal API data into object
            var adv defs.AgentAdvert
            err = json.Unmarshal([]byte(resp.Node.Value), &adv)
            common.FailOnError(err, "Failed to retrieve data from server")

            // We want to use the TTLDuration field from etcd for simplicity
            adv.Expires = resp.Node.TTLDuration()

            // The etcd key should always match the inner JSON, so let's panic (for now) if this ever happens
            if uuid != adv.Uuid {
                panic("UUID in etcd does not match inner JSON text")
            }

            ret_adv = append(ret_adv, adv)

        }
    }

    return ret_adv
}

func (etcddb EtcdDB) RemoveGroupFile() {
}
