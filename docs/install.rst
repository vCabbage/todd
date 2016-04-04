Installing and/or Running ToDD
================================

First, make sure the following software is installed and correctly configured for your platform:

- Go (1.6 is the version tested for this documentation)
- Make sure the "bin" directory in your GOPATH is also added to your "PATH"
- Git

ToDD can be run in a number of ways. The primarily

The makefile provided in the repo can be used to easily install

.. code-block:: text
    go get -d github.com/Mierdin/todd

At this point, you may get an error along the lines of "no buildable GO source files in...". Ignore this error, you should still be able to install ToDD.

Navigate to the directory where Go would have downloaded ToDD. As an example:

.. code-block:: text
    cd $GOPATH/src/github.com/Mierdin/todd

Finally, compile and install the binaries:

.. code-block:: text
    make
    sudo make install

Docker
----------
If you instead wish to run ToDD inside a Docker container, you can pull the current image from Dockerhub:

.. code-block:: text
    mierdin@todd-1:~$ docker pull mierdin/todd
    mierdin@todd-1:~$ docker run --rm mierdin/todd todd -h                        
    NAME:
       todd - A highly extensible framework for distributed testing on demand

    USAGE:
       todd [global options] command [command options] [arguments...]

    VERSION:
       v0.1.0
    ......

A Dockerfile for running any ToDD component (server/agent/client) is provided in the repository if you wish to build the image yourself. This Dockerfile is what's used to automatically build the Docker image within Dockerhub.

Vagrant
----------
There is also a provided vagrantfile in the repo. This is not something you should use to actually run ToDD in production, but it is handy to get a quick server stood up, alongside all of the other dependencies like a database.