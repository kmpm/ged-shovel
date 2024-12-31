# EDDN to Nats Proxy
Receives messages from EDDN and republish them on a given NATS.io server



- https://github.com/EDCD/EDDN/blob/live/docs/Developers.md


## Docker
```powershell
docker build -t ged-enp . 
docker run -d --name enp --network nats --rm ged-enp


go install github.com/kevin-cantwell/zlib/cmd/zlib@latest

nats sub --translate "zlib -d"  "eddn.journal.1"
nats sub --translate "zlib -d"  "eddn.>"
```