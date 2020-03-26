# Selfor
Selfor (pronounced like _Sulfur_) is a **Sel**ective TCP **For**warder. Connections to selfor are forwarded to a host defined in a Redis instance.

## Motivation
SSH cannot be load balanced on the basis of hostname due to the nature of the protocol, but with selfor, you can expose multiple servers inside a firewall to the outside world with a single open port. For every source IP address, a destination is mapped to the forwarded, which may be chosen externally by simply setting a value in Redis.

## Redis
Each key in redis should be the source IP as the key and the value as a hashmap with `d` as the destination and `i` as extra logging info.
```
HSET 103.4.1.33 d 10.100.100.100:22
HSET 103.4.1.33 i myuser
EXPIRE 103.4.1.3 3600
```
Such calls may be made by an externally authenticated application (e.g. by a web interface). An example application in PHP (without any authentication) is provided in this repo.

## PROXY protocol
Connections from nginx stream module are supported with remote identification with PROXY protocol

