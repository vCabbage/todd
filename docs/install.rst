Installing and/or Running ToDD
================================

ToDD can be run in a number of ways.

Makefile
----------
The makefile provided in the repo can be used to easily install


Docker
----------
A Dockerfile for running any ToDD component (server/agent/client) is provided in the repository, and this Dockerfile is also used in the ToDD CI pipeline to automatically build a docker image and push to Dockerhub. Thus, you can easily retrieve the latest version of this image by performing a ``docker pull``:

.. code-block:: text
    mierdin@todd-1:~$ docker pull mierdin/todd

Vagrant
----------
There is also a provided vagrantfile in the repo. This is not something you should use to actually run ToDD in production, but it is handy to get a quick server stood up, alongside all of the other dependencies like a database.