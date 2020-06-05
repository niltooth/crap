# ItsATrap!
A high performance snmp trap reciver. Traps are converted into json and stored as jsonb in PostgreSQL. If you enable TimescaleDB in PostgreSQL will you have a highly scalable timeseries database with the full power of PostgreSQL at your hands. That includes full json support, joins, etc.

## Performance
On my limited home hardware I was able to process 20k traps per second. 
Data is written to an in memory buffer to allow for bulk sql inserts via COPY. Testing on the lan I have been able to flush 400k traps to the db in about 3 seconds. This could probably be further optimized by using binary copy.
## TODO

- Add option to throw all messages through nats. this will allow multiple receivers.
 
 `1. trap -> 2. encode to json -> 3. throw to nats <- 4. receive from nats -> 5. buffer -> 6. flush to db`
 
 The process should be able to run in 3 modes.
	1. Trap to buffer to db. (all in one without nats)
        2. Trap to nats (a single process)
        3. From nats to buffer to db (a single process)
  	2 & 3 would do the work of 1 with nats in the middle

- Add index option
