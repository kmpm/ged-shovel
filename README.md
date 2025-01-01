# Go Elite:Dangerous - Shovel
Receives messages from EDDN and republish them on a given NATS.io server,
written in Go



- https://github.com/EDCD/EDDN/blob/live/docs/Developers.md


## Docker
```powershell
docker build -t ged-shovel .
docker run -d --name shovel --network nats --rm ged-shovel


go install github.com/kevin-cantwell/zlib/cmd/zlib@latest

nats sub --translate "zlib -d"  "eddn.journal.1"
nats sub --translate "zlib -d"  "eddn.>"
```
