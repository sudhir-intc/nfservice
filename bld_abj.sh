set -x
docker volume prune
docker system prune -a
source env_abj.sh
docker images | grep -e nf. -e REPO
#docker rmi -f <nf1 container id> 

make build
make build-docker

