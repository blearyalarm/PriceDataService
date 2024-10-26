docker build -t gomono-protos:latest . -f etc/script/Dockerfile.protos
id=$(docker create gomono-protos:latest)
docker cp $id:/app/gen/. ./.gen/protos
docker rm -v $id
docker rmi gomono-protos:latest
