Working with the ToDD Client
================================

ToDD comes with an easy-to-use CLI client. Use the ``--help`` flag to print some useful help output:

.. code-block:: text
   :emphasize-lines: 1

    mierdin@todd-1:~$ todd --help
    NAME:
       todd - A highly extensible framework for distributed testing on demand

    USAGE:
       todd [global options] command [command options] [arguments...]

    VERSION:
       v0.1.0

    COMMANDS:
       agents   Show ToDD agent information
       create   Create ToDD object (group, testrun, etc.)
       delete   Delete ToDD object
       groups   Show current agent-to-group mappings
       objects  Show information about installed group objects
       run      Execute an already uploaded testrun object
       help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
       -H, --host "localhost"   ToDD server hostname
       -P, --port "8080"        ToDD server API port
       --help, -h           show help
       --version, -v        print the version


Agents
----------

Use the ``todd agents`` command to display information about the agents currently known to the ToDD server.

.. code-block:: text
    mierdin@todd-1:~$ todd agents --help
    NAME:
       temp-client agents - Show ToDD agent information

    USAGE:
       temp-client agents [arguments...]

You can run this command on it's own to display a summary of all agents, their facts, collectors, etc:

  < example >

Or, you could append an agent UUID to this command to see detailed information about that agent, such as the facts that it is reporting:

  < example >

Create
----------

Run "todd create"

Delete
----------

Run "todd delete"

Groups
----------

Run "todd create"

Objects
----------

Run "todd create"

Run
----------

Run "todd create"

Show optional arguments