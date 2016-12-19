docker swarm init --advertise-addr eth0:2377 --listen-addr eth0:2377

curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/managersJson > /tmp/managers.json
curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/workersJson > /tmp/workers.json

download "infrakit"
for i in $(seq 1 60); do /tmp/infrakit group commit /tmp/managers.json && break || sleep 1; done
for i in $(seq 1 60); do /tmp/infrakit group commit /tmp/workers.json && break || sleep 1; done
