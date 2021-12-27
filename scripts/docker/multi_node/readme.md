
# how to use

step1. change `crypto_config[0].count=n` in file `chainmaker-go/tools/chainmaker-cryptogen/config/crypto_config_template.yml`


`crypto_config_template.yml`
```yaml
crypto_config:
  - domain: chainmaker.org
    host_name: wx-org
    count: 7                # change this what you want, example is 7 node
```

step2. change `IMAGE` in file create_docker_compose_yml.sh
`create_docker_compose_yml.sh`
```yaml
P2P_PORT=$1
RPC_PORT=$2
NODE_COUNT=$3
CONFIG_DIR=$4
SERVER_COUNT=$5
IMAGE="chainmakerofficial/chainmaker:v2.1.0" # change this
```

step3. prepare 
```sh
cd chainmaker-go/script 
./prepare.sh 4 1 11331 12331, example is 4 consensus node 3 common node
```

step4. change and execute 
```sh
cd chainmaker-go
cp -rf build/config scripts/docker/multi_node/
cd scripts/docker/multi_node
# change ip what you want
sed -i "s%127.0.0.1%192.168.1.35%g" config/node*/chainmaker.yml
./create_docker_compose_yml.sh 11331 12331 7 ./config 4
```

step5. run docker with compose
```sh
docker-compose -f docker-compose1.yml up -d
docker-compose -f docker-compose5.yml up -d
# or 
./start.sh docker-compose1.yml
./start.sh docker-compose5.yml
```

tips: stop docker with compose
```sh
docker-compose -f docker-compose1.yml down
docker-compose -f docker-compose5.yml down
# or
./stop.sh docker-compose1.yml
./stop.sh docker-compose5.yml
```
