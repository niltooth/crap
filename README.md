# crap
A high performance snmp trap to nats proxy. 

Receives snmp traps, encodes them as json and publishes them to a nats subject.

It only starts here.. continue by enabling jetstream, adding streams and writing consumers.

This system should be able to scale from a single embedded instance on a raspberry pi to a very large cluster. 

Messages can be reliably delivered once they have been received by ```crap```

## Features
- Encode snmp traps (v1,v2 & v3) into JSON
- Internal stats sent to nats


Example config can be found in the base of the github repo
