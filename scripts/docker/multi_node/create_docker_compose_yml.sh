
P2P_PORT=$1
RPC_PORT=$2
NODE_COUNT=$3
CONFIG_DIR=$4
SERVER_COUNT=$5
IMAGE="chainmakerofficial/chainmaker:v2.2.1"

CURRENT_PATH=$(pwd)
CONFIG_FILE="docker-compose"
TEMPLATE_FILE="tpl_docker-compose_services.yml"

function show_help() {
    echo "Usage:  "
    echo "  create_yml.sh P2P_PORT RPC_PORT NODE_COUNT CONFIG_DIR SERVER_NODE_COUNT"
    echo "    P2P_PORT:          peer to peer connect"
    echo "    RPC_PORT:          sdk to peer connect"
    echo "    NODE_COUNT:        total node count"
    echo "    CONFIG_DIR:        all node config path, relative or absolute"
    echo "    SERVER_NODE_COUNT: default:100, number of nodes per server"
    echo ""
    echo "    eg: ./create_docker_compose_yml.sh 11301 12301 4 ./config : 4 nodes in 1 machine"
    echo "    eg: ./create_docker_compose_yml.sh 11301 12301 16 ./config 2 : 4 nodes in 8 machine, 2 nodes per machine"
    echo "    eg: ./create_docker_compose_yml.sh 11301 12301 4 /data/workspace/chainmaker-go/build/config 4"
}
if [ ! $# -eq 2 ] && [ ! $# -eq 3 ] && [ ! $# -eq 4 ] && [ ! $# -eq 5 ]; then
    echo "invalid params"
    show_help
    exit 1
fi

function xsed() {
    system=$(uname)

    if [ "${system}" = "Linux" ]; then
        sed -i "$@"
    else
        sed -i '' "$@"
    fi
}

function check_params() {
    if  [[ ! -n $P2P_PORT ]] ;then
        show_help
        exit 1
    fi

    if  [ ${P2P_PORT} -ge 60000 ] || [ ${P2P_PORT} -le 10000 ];then
        echo "P2P_PORT should >=10000 && <=60000"
        show_help
        exit 1
    fi

    if  [[ ! -n $RPC_PORT ]] ;then
        show_help
        exit 1
    fi

    if  [ ${RPC_PORT} -ge 60000 ] || [ ${RPC_PORT} -le 10000 ];then
        echo "RPC_PORT should >=10000 && <=60000"
        show_help
        exit 1
    fi

    if  [[ ! -n $NODE_COUNT ]] ;then
        show_help
        exit 1
    fi

    if  [[ ! -n $SERVER_COUNT ]] ;then
        SERVER_COUNT=100
    fi

    if  [[ ! -n $CONFIG_DIR ]] ;then
        CONFIG_DIR="../../../build/config"
    fi
}

function xsed() {
    system=$(uname)

    if [ "${system}" = "Linux" ]; then
        sed -i "$@"
    else
        sed -i '' "$@"
    fi
}

function generate_yml_file() {
  tmp_file="${TEMPLATE_FILE}.tmp"
  current_config_file=""
  for ((k = 1; k < $NODE_COUNT + 1; k = k + 1)); do
    surplus=$(( $(($k - 1)) % $SERVER_COUNT ))
    if [ $surplus -eq 0 ]; then
      current_config_file="${CONFIG_FILE}${k}.yml"
      rm -rf $current_config_file
      echo "generate $current_config_file"
      echo -e "version: '3'\n"  >> $current_config_file
      echo -e "services:"  >> $current_config_file
    fi
    if [ ! -f $tmp_file ];then
      cp $TEMPLATE_FILE $tmp_file
    fi
    node_config_dir="${CONFIG_DIR}/node${k}"
    xsed "s%{config_dir}%${node_config_dir}%g" $tmp_file
    xsed "s%{id}%${k}%g" $tmp_file
    xsed "s%{image}%${IMAGE}%g" $tmp_file
    xsed "s%{rpc_port}%${RPC_PORT}%g" $tmp_file
    xsed "s%{p2p_port}%${P2P_PORT}%g" $tmp_file
    cat $tmp_file >> $current_config_file
    rm -f $tmp_file
    P2P_PORT=$(($P2P_PORT+1))
    RPC_PORT=$(($RPC_PORT+1))
  done

}
check_params
generate_yml_file
