ToDD Testlets
================================

Tests in ToDD are powered by something called "testlets". These are executable files (i.e. scripts, binaries) that accept a standard set of input, run a test application, and provide a standard set of output containing metrics from that test.

This allows the user to simply "use" this application, and specify which agents run this application. All of the complicated stuff with respect to sending arguments to the underlying testing application as well as parsing the output, is performed inside the testlet.

Input
----------
A group object defines the rules that will place an agent into the referenced group name.


