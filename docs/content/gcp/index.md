Open the Google Cloud Console. 
Run:
gcloud deployment-manager deployments create docker --config https://storage.googleapis.com/docker-template/swarm.py --properties managerCount=3,workerCount=1



To have prompts:
curl -sSL http://get.docker-gcp.com/ | sh

