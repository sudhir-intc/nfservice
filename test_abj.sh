set -x
curl http://localhost:8060/nf1 0
sleep 10
curl http://localhost:8060/nf2loc
sleep 10
curl http://localhost:8060/nf1  1
sleep 10
#wget http://localhost:8060/nf2loc

curl http://localhost:8060/nf2loc
wget http://localhost:8060/nf2loc


exit 0
docker stop `docker ps -a -q`
docker start `docker images -a -q`

docker images | grep -e nf. -e REPO
docker run -dit <Image ID>

docker ps 
docker exec -it <Container ID> /bin/bash

ping -c2 172.17.0.1
ping -c2 172.17.0.2
ping -c2 172.17.0.3
route -n
