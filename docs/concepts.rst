ToDD Concepts
================================

Before exploring the details of how ToDD actually works, it's important to first spend some
time understanding a few fundamental concepts.

Groups
------
In order to solve the problems that ToDD was created to solve, it was necessary to consider scale
from the beginning. In ToDD, testing is performed much more like "cattle" than "pets". As a result,
tests are not performed with discrete endpoints, such as ToDD agents, but rather with "groups".

.. NOTE::

   At it's lowest level, testing in ToDD is done either from a group of agents, or between groups of
   agents. It does not deal with individual agents. This is a very important concept to remember,
   as all other concepts build on this.

You should still consider a "group" as a singular logical endpoint for a test, even though there
may be multiple actual endpoints within that group.

Groups can contain as few as a single ToDD agent, or as many as tens or even hundreds. The idea
is to detach the concept of testing from discrete nodes or IP addresses, and provide a grouping
mechanism that allows the ToDD user to adjust the scale or power of a test without fundamentally
changing its parameters.

This concept shows its true power when you compare two instances of the same test being run. In the
diagram below, we have two instances of the same test. In the first instance, only one agent is
present in each group. In the second, several additional agents are present in each group:

.. image:: images/groupcompare.png

Note that both the group configuration, as well as the configuration of the test itself, are identical
between these two test instances. This is because the mechanism by which you describe groups in ToDD
allows you to anticipate different "test topologies" ahead of time. Tests are performed between groups,
regardless of what the group happened to contain at test time.

For instance, the group "datacenter" might be configured to include all agents that belong to a
certain subnet. Once this configuration is in place, scaling this group to include additional
agents is a simple matter of spinning up additional agents that belong to that subnet.

.. NOTE::

   See the "Groups" section of the `Objects Documentation <using/objects.html>`_  for more info on
   what parameters can be used to assign agents to groups.

So, at this point, you can probably think of groups best as a scaling mechanism. ToDD assumes you have
other tools, such as Kubernetes, for automating this, and tries only to allow you to describe how that
might look so that ToDD can automatically use resources you've provided for it.
