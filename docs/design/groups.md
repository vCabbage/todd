ToDD Agent Grouping
====

# Outline

1. Introduction to Facts
2. Facts as JSON
3. Extending Facts
    a. Fact Exchange
4. Using Facts to Group Agents



Grouping is a discrete (is this the right word for run-once, on demand?), (synchronous?) act within ToDD. The role of the operator is to put the right agents into place, and registering with the ToDD server, and once things are working, run this grouping mechanism. Agents are grouped, and those groupings are stored.

What happens if an agent goes offline, even temporarily? Will the agent be removed from the group it was assigned to?

As a result of the last question, when is the grouping mechanism run? Is it run prior to a test? Maybe we should support a group only mechanism, which doesn't run the tests, but when the test is run, the grouping is run next. This way, the grouping doesn't have to run separately from the testing if you don't want, but you can also run the grouping mechanism first to see how the agents would be grouped if you were to run the test.

Where is the grouping information assigned? Presumably the database.

What is the name of the group "file"? Need to come up with a name for this so you can refer to it easily.


# Roadmap

The first version of the grouping mechanism will be very static, using statements that operate in a very specific way. It will not be very flexible, and will assume that you're only using the default fact collectors, since they're statically written to look for certain facts. This was done in order to 

However, in the very near future, the grouping mechanism will be refactored to be much more flexible, to allow for more generic language and logic to match against things in a given fact, regardless of how it's built. This will allow users to develop their own fact collectors, and immediately write grouping statements to group agents accordingly.

This approach was chosen to get a basic grouping model (that will likely work for >90% of users) into ToDD quickly, allowing us to move faster towards the actual testing development. 



## Fact Filters

Fact filters contain logic relevant to the particular fact key they're referring to.

fact filters should have a "name" property, and both this property and the filename of the fact filter should be named according to the factset they intend to target

TODO
If you create a fact collector, you should also create a fact filter?? (how?)
Need to consider renaming these to be more descriptive of what they actually do