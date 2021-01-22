# ItsATrap!
A high performance snmp trap receiver. Traps are converted into json and stored as jsonb in PostgreSQL. 
If you enable extensions like TimescaleDB/Citus or run Distributed SQL like Yugabyte, Aurora or Cockroach will you have a highly scalable timeseries database that can co-exist with your other data in the same cluster. 
This includes full json support with joins, indexing, etc.

## Features
- Encode snmp traps (v1,v2 & v3) into JSON
- Bulk load traps into any postgresql client compatible database (postgresql, cochroachdb, yugabyte, etc)
- Uses COPY instead of insert for better performance
- Publish traps to nats cluster for "fan out" or load balancing.
- 3 modes. 
    - db-only: write traps only to the db
    - nats-only: write traps only to a nats subject
    - hybrid: write traps to both nats and the db
## Performance
On my limited home hardware I was able to process roughly 50k traps per second. 
Data is written to an in memory buffer to allow for bulk sql inserts via COPY. Testing on the lan I have been able to flush about 1 million traps to the db in about 1 second. This could probably be further optimized by using binary copy.

## TODO
- Code cleanup and documentation.
- Test performance on databases other than vanilla PostgreSQL.
- If additional outputs are needed, move to a plugin based design.
