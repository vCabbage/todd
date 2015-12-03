ToDD Facts LLD Documentation
====

## Outline

1. Introduction to Facts
2. Facts as JSON
3. Extending Facts
    a. Fact Exchange
4. Using Facts to Group Agents


# Introduction to Facts

The purpose of facts in ToDD is to establish a robust system of identification for groups of ToDD agents. Test runs will identify which groups of agents are used in what way during a test, and as a result, the user needs to be able to logically group agents together so that they can take place in this operation.

ToDD allows this by use of facts. Facts should be thought of loosely as "properties of an agent and it's environment". At runtime, ToDD uses logic imparted by the user to group agents based on these facts. (TODO link to grouping doc for this)

## Facts as JSON

The JSON describing a single set of facts must contain a single key-value pair. The key should be a string that uniquely represents this fact set to the server. The value for this key can be just about anything - this is specific to the implementation of this fact set.

    {
      "Interfaces": [
        {
          "Name": "lo0",
          "HwAddr": "",
          "IPv4Addrs": [
            "127.0.0.1",
            "127.94.0.3",
            "127.94.0.1"
          ],
          "IPv6Addrs": [
            "::1",
            "fe80::1"
          ]
        },
        {
          "Name": "en4",
          "HwAddr": "0c:4d:e9:d5:28:45",
          "IPv4Addrs": [
            "10.12.0.193"
          ],
          "IPv6Addrs": [
            "fe80::e4d:e9ff:fed5:2845"
          ]
        }
      ]
    }

This is actually a single fact - namely "Interfaces". The internals of this fact are actually not relevant during the collection process - the corresponding value for the key "Interfaces" will not be checked until it's time to filter based on facts, such as during agent grouping.

## Extending Facts

As mentioned in the previous section, facts are represented by a key/value construct, represented in JSON.

ToDD obtains these JSON structures by executing standalone binaries or scripts on each agent, called "fact collectors", or just "collectors". Collectors are responsible for gathering a single fact (such as the "Interfaces" fact shown in the example above), and write it to stdout as JSON.

As a result, each executable or binary needs to be self-sufficient, as it will be run like so:

    ./getinterfaces

Each collector is responsible for one and only one fact. If the resulting JSON

The ToDD project comes with a few fact collectors by default (see the facts/collectors directory), but because of this simple interface, it's fairly easy to write your own. If you wish to do this, just follow these three rules:

- Write your collector to be a totally self-sufficient executable. This could be a compiled binary, or a script with the appropriate shebangs to indicate an interpreter (the default collectors that come with ToDD are written in Python). ToDD will distribute these files to each agent, set the execute permissions on them, and attempt to run them as-is.
- The collector needs to implement a JSON schema that follows the standards described in the previous section, and writes it to stdout.
- Place your collector file in the directory "facts/collectors" within the ToDD repository. 

When you compile ToDD using the providing Makefile, these collectors will be embedded within the resulting binary for the ToDD server. You don't need to worry about moving these files anywhere else on the system.

TODO Mention here to look at the fact filters section, as you'll likely need to provide a fact filter for each fact collector

### Distribution of Fact Collectors

Since fact-gathering is built to be extensible, the server needs to have a mechanism that ensures all registered agents have all of the appropriate collectors installed.

To accomplish this ToDD implements a simple exchange mechanism that piggybacks on the normal agent advertisement. In addition to advertising information about itself, an agent will also send a list of installed fact collectors, and the calculated SHA256 hash of each collector file.

    ---- FACT CAPABILITY ADVERTISEMENT FROM AGENT TO SERVER ----

    {

      ... rest of agent advertisement omitted for brevity...

      "FactCollectors": {
        "get_hostname": "19c531c02a1f5684cfae826d063deac0520378fbf1034600d2b56f80bad41104",
        "get_interfaces": "9f216c4689289fd2924af361206c085f78f6d25a878717aa925a7bf4b0088d86"
      }
    }

The ToDD server maintains it's own authoritative list of which fact collectors are active, and will look at each agent advertisement to ensure that all of the required files (with correct hashes) are being advertised by each agent.

In the event that an agent is missing a collector file, or perhaps has a collector with a different hash than what was calculated by the server, the server will issue a remediation notice, so the agent can download collector files and become compliant. This message is sent over the message queue - directly to the offending agent - and is encoded in JSON:

    ---- FACT REMEDIATION NOTICE FROM SERVER TO AGENT ----

    {
      "FactRemediation": [
        "http://192.168.0.101:8090/core_network_interfaces_1_0",
        "http://192.168.0.101:8090/core_hostname_1_0"
      ]
    }

This is a simple list of download URLs that the agent should retrieve to become compliant. The agent should overwrite any files that already exist, and each file will have the execute permission set on it, so that they can be executed during the next advertisement.

When an advertisement results in a remediation notice, the server will not add the agent to the local registry. The agent will only be added when it provides a compliant list of fact collectors. This means that an agent will not be available for testing until it reports a satisfactory collector list to the server in its advertisements.

## Using Facts to Group Agents

This is in a separate doc "groups.md"

## TODO

1. What happens if we have fact naming collisions? Do we just disable the agents? Basically if someone wrote a fact collector poorly and the name of the fact set in the JSON is the same as one that already exists?

