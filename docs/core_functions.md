## Modeling Control Plane and Data Plane Entities and Relations using onos-topo


## Core functionality
onos-p4-sdk implements the following core functions to make interaction of SDN apps and programmable devices using P4Runtime possible.
The major core functionality are listed as follows:
### Southbound
It implements a P4Runtime southbound client interface and a connection
manager to maintain list of connections that are established between the app and devices. It also provides the capability
to detect connection failures and generates connection events
### Reconciliation loop Controllers
- *Target Controller*:  This controller is responsible to add/remove a connection to a P4 programmable target when the corresponding entity is
  added/removed from onos-topo or connection events (e.g. connection failure).
- *Node Controller*: This controller is responsible to create CONTROLLER entities for each instance of an SDN application.  
  It also implements a lease mechanism to renew the lease expiration aspect on local CONTROLLER entities delete CONTROLLER entities if their lease expired (e.g. deleting an application pod)
- *Mastership Controller*: This controller is responsible for preparing and sending MasterArbitrationUpdate request after a connection is made to the target and processing the response that it receives from the device. After processing
  the response, it updates mastership state aspect for P4 programmable entity in onos-topo.
- *Connection Controller*: This controller is responsible to Create/Delete CONTROL
  relations between P4 programmable entities (e.g. a SWITCH entity) and CONTROLLER entity. 