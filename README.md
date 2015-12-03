ToDD
====
ToDD stands for "Testing on Demand: Distributed!". ToDD is an extensible framework for providing natively distributed testing on demand.

Modern testing frameworks do a great job of performing their test suite, but often don't consider how to scale-up their testing software until much later, if they do at all. As a result, each testing application must independently consider how to scale beyond the limitations of a single host, and implement this in a specific way. That's a lot of duplicated time and code!

ToDD provides a set of tools that remove the need for a testing application to worry about the complexities of operating in a distributed manner. Using these tools, an operations team can deploy any testing application within ToDD, and take advantage of the benefits of a fully distributed model. Because of this distributed model, ToDD helps to guarantee test diversity - instead of simply testing between two points, which can be fairly isolated, ToDD helps to encourage utilization of more of your infrastructure, so that any failures or shortcomings can be discovered more quickly.

In addition, ToDD treats post-test analytics as a first-class citizen. There is an incredible amount of value in being able to definitively see the impact of an infrastructure change, by first taking a baseline, and then running the same test after a change. ToDD makes this super easy to do, and brings a whole new level of operational awareness to deploying applications.

ToDD is designed to be totally stateless. All state is moved into the database layer. For instance, if ToDD is sitting on top of etcd, then etcd holds all state information, such as what agents are currently registered with the server.

# Quickstart

The best way to run ToDD is with Docker. The Makefile found in this repository heavily favors a Docker-based build and run environment. First, compile ToDD with make:
    
    make

Then you can run all required components within Docker containers with:

    make run

However, if you wish to compile and run ToDD directly, you can also run this to compile and install ToDD binaries on the local OS:

    make install

Assuming you started ToDD with "make run", you should see three containers running:

    ~$ docker ps                                                                                                            
    CONTAINER ID        IMAGE                   COMMAND                  CREATED             STATUS              PORTS                                             NAMES
    0159b39a634e        mierdin/todd            "todd-agent"             46 minutes ago      Up 46 minutes                                                         todd-agent
    67a60abfb39b        mierdin/todd            "todd-server"            46 minutes ago      Up 46 minutes       0.0.0.0:8080->8080/tcp                            todd-server
    375adfb27e86        rabbitmq:3-management   "/docker-entrypoint.s"   46 minutes ago      Up 46 minutes       0.0.0.0:5672->5672/tcp, 0.0.0.0:8090->15672/tcp   todd-rabbit

You can run todd-client within it's own container to query agent status (this example's docker host is running at 192.168.59.103)

    ~$ docker run mierdin/todd todd-client --host="192.168.59.103" --port="8080" agent
    -------------------------------------------------------
    Agent UUID:  19ebbed51d1891a1ffa6cdd65e203ee322ebf35d924ced40cb1c80a6dd106ee5
    Enabled:  false
    Last Seen:  15.722431ms
    Hostname:  0159b39a634e
    -------------------------------------------------------

# Features

ToDD provides the following features:

- 100% user-defined, YAML-based application manifests, so that any application can be distributed and run within ToDD.
- Post-test analytics, made possible by a time-series database. See the impact of infrastructure changes very easily by looking at two or more test runs.
- Extensible grouping mechanism - be as granular as you want to be regarding the systems participating in a test.
- Diverse target types - can test against other ToDD agents, or uncontrolled systems.

# Architecture

Summary of architecture in lieu of formal documentation:

- All message queue calls are centralized, and made as generic as possible. This will allow us to move to a more pluggable message queue abstraction down the road

# Disclaimer

Contributors to ToDD are doing so as independent contributors. No contribution to this project should contain any intellectual property of any contributor's employer (past or present).

(This section really needs more - maybe a terms of contribution?)