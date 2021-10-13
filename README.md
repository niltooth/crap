# crap
A high performance snmp trap to nats proxy. 
Receives snmp traps, encodes them as json or protobuf and publishes them to a nats subject.

This tool is intended to be a building block for other systems. Not a full blown alerting engine with a trap receiver.

It only starts here.. continue by enabling jetstream, adding streams and writing consumers.

This system should be able to scale from a single embedded instance on a raspberry pi to a very large cluster. 

Messages can be reliably delivered once they have been received by ```crap```

## Status
This is mostly stable, but reporting issues will help a lot. 
## Deployments
- Local. By running local the machine can basically send its traps over nats. This removes complexity with configuration management, h/a and load balancing
- Cloud/Datacenter with single node or multiple node
- Edge/Branch: If message reliability is of concern you may want to deploy an instance to edge/branch locations. This will let messages get buffered on the LAN and reliably sent upstream.   

## Features
- Encode snmp traps (v1,v2 & v3) into Protocol Buffers or JSON
- Internal stats sent to nats


Example config can be found in the base of the github repo
