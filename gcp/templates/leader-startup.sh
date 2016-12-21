docker swarm init --advertise-addr eth0:2377 --listen-addr eth0:2377

configs=/infrakit/configs
mkdir -p $configs

curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/managersJson > $configs/managers.json
curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/workersJson > $configs/workers.json

infrakit="docker run --rm $local_store $image infrakit"
for i in $(seq 1 60); do $infrakit group commit /root/.infrakit/configs/managers.json && break || sleep 1; done
for i in $(seq 1 60); do $infrakit group commit /root/.infrakit/configs/workers.json && break || sleep 1; done
