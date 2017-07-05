
# Summary
-  This is an attempt to the coding challenge:https://github.com/InsightDataScience/anomaly_detection/blob/master/README.md

## Dependencies
- The project needs Go (golang) compiler of version 1.7.4
- There are no external dependencies and it uses all the standard package which are part of GO, and all the code is self contained

## Compilation instructions
- I am using go 1.7.4, and it should work with other later version too (Tested it on go version 1.6.2 too)
- Once go is installed do the following
    - setenv GOPATH <location for go development> (fully qualified path)
    - `cd $GOPATH`
    - `mkdir src`
    - `cd $GOPATH/src`
    - `go get github.com/apganesh/anomaly_detection`
        (This should download the whole repo under $GOPATH/src/github.com/apganesh/anomaly_detection)

    - To run the tests you can "cd" to 
         - cd `$GOPATH/src/github.com/apganesh/anomaly_detection/insight_testsuite `
         - `./run_tests.sh`
         - File under `$GOPATH/src/github.com/apganesh/anomaly_detection/run.sh` has been modified to build the executable "anomaly_detector"

    - To compile it manually
        - `cd $GOPATH/src/github.com/apganesh/anomaly_detection/`
        - execute ./run.sh
        - This should create an executable called "anomaly_detector" under `$GOPATH/src/github.com/apganesh/anomaly_detection/src/anomaly_detector`
        - We can execute the tests from command line:
        `$GOPATH/src/github.com/apganesh/anomaly_detection/src/anomaly_detector <path to batch_log.json> <path to stream_log.json> <path to flagged_purchases.json>`

# Toplevel function called from main function:

- ReadBatchFile
        - read the batch file and creates a graph with nodes/vertices which represents a social graph.
        - all the purchase information is recorded in Purchase object which includes the unique timesequence at which the purchase is made along with amount and the id of the member
        - for all the relation events like "befriend", "unfriend", we update the graph and mark the related nodes as "dirty", which will be lazily processed while reading the stream file 
- ReadStreamFile
        - This function parses the stream_log.json file and processes the events in "active" way.  
        - When ever "purchase" event comes in, we check if the node belonging to id, is "dirty" or not.
            - a node becomes dirty if any of its connected friends with in the 'D' range has been affected by a "befriend" or "unfriend"
            - depending on "dirty" or not it either returns the cached mean and standard deviation, or uses a modified k-way merge to recalculate the most recent (T) transactions from its friends network
            - once the mean and standard deviation infomation is calculated / (or from cache), we decide if the purchase is anomaly or not, and flag it as an entry in flagged_purchases.json


# utils
- AnamolyDetection struct encapsulates the following:
    - graph object which represents the social network as an adjacency list
    - transactions (purchases) made by each user in the social network.  We keep only the latest "T" purchases made. 
    - It also stores some of the global values use all across the code such as
        - D,T provided by the batch_log.json, 
        - globalsequence counter which keeps track of the order in which the purchase's were made

    - This object provides all the high level API calls for 
        - add an edge between two members
        - remove an edge between two members
        - list all the connected members for a given id with in distance of 'D'
        - add a purchase and check for anomaly
    - This object is responsible for reading the batch_log.json and stream_log.json files


- PurchaseData keeps track of the following information for each "id" (node)
    - Its current friends in the network within distance 'D'
    - Its latest 'T' purchases in a sorted manner
    - Its latest 'T' purchases from its friends in the network
    - Cached values of sum/sum^2 of all the 'T' latest purchases from its connected friends
    - Also the dirty flag, which represents if this node is affected by any "befriend" or "unfriend" events

    - This object also provides the following API
        - update the member's latest purchases by keeping only the latest 'T' transactions
        - update the member's network latest 'T' purchases using a modified k-way merge function which picks the latest 'T' transactions from its friends' network in an optimal manner
        - updates and caches the sum and sum^2 of all its friends' 'T' purchases which will be used for calculating mean and standard deviation.

- Graph object proviedes the API for manipulating the social network.  APIs are provided for operations like addVertex, addEdge, removeEdge, and getFriends within a distance of 'D'.

# priorityqueue
- This module provides a genric priorityqueue implementation on top of the heap datastructure provided by the standard library in GO.


# Experimentation

- Various experimentations on how to achieve the optimal run time.

## Idea 1
- Initially the problem seemed to be trivial, but once started working with huge test cases with a huget network, figured out that doing a merge of all the purchases in the friends's network, seems to be compute intesive.  Merging T * number of friends and find the Mean and StandardDeviation for every 'purchase' seems to be and overkill.  Used the heap datastructure to collate all the purchases and heapified it find the top 'T' latest purchases.

## Idea 2

- Second experiment was to keep track of the node's status "dirty" or not.  The "dirty" flag is set on all the nodes which were affected by "unfriend" or "befriend".  As the network grows we might have isolated nodes which were not affected at all, and we dont have to recomupte the latest 'T' purchases in its network. The cached sum and sum^2 can be used to calculate the mean and standard deviations.
    - Merging of all the connected friends purchases to produce 'T' latest purchases is as follows:
            - As the purchases made by each member is stroed in an ascending order of timestamp, we process each entry in reverse order which adding it the the minPQ.
            - We sort the minPQ using a globalseq number which represents the order in which the "purchase" events occured.
    - The algorithm goes like this:
             - We initialize a min priority_queue (minPQ)
             - Initialize a processing list with all the friends information
             - For each friend's purchase in reverse order of its original purchases
                - Keep adding the purchase to the minPQ till we reach the size of minPQ to 'T'
                    - Continue till we have 'T' entries in the minPQ
                - We keep track of the indices of each friends' purchase list
                - If for an entry has a sequence number > top items sequcence number
                    - Pop the top item and remove the top items' ID from the processing list (as this friends' prior purchased will not make it into minPQ we dont have to process it)
                    - Add the new entry into the minPQ.
                - Else 
                    - We can safely remove this friend from the processing list

    - The second experiment provided better runtime when compared to the first idea.  More optimization can be done on with respect to recomputing the members connected friends, when ever "befriend" or "unfriend" events occur.


