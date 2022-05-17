#!/usr/bin/env bash
#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

NODE_CNT=$1
CHAIN_CNT=$2
P2P_PORT=$3
RPC_PORT=$4

CURRENT_PATH=$(pwd)
PROJECT_PATH=$(dirname "${CURRENT_PATH}")
BUILD_PATH=${PROJECT_PATH}/build
CONFIG_TPL_PATH=${PROJECT_PATH}/config/config_tpl_pwk
BUILD_CRYPTO_CONFIG_PATH=${BUILD_PATH}/crypto-config
BUILD_CONFIG_PATH=${BUILD_PATH}/config

CRYPTOGEN_TOOL_PATH=${PROJECT_PATH}/tools/chainmaker-cryptogen
CRYPTOGEN_TOOL_BIN=${CRYPTOGEN_TOOL_PATH}/bin/chainmaker-cryptogen
CRYPTOGEN_TOOL_CONF=${CRYPTOGEN_TOOL_PATH}/config/pwk_config_template.yml
#CRYPTOGEN_TOOL_PKCS11_KEYS=${CRYPTOGEN_TOOL_PATH}/config/pkcs11_keys.yml

VERSION=v2.2.1

BC_YML_TRUST_ROOT_LINE=$(awk '/trust roots list start/{print NR}' ${CONFIG_TPL_PATH}/chainconfig/bc_4_7.tpl)
BC_YML_TRUST_ROOT_LINE_END=$(awk '/trust roots list end/{print NR}' ${CONFIG_TPL_PATH}/chainconfig/bc_4_7.tpl)

function show_help() {
    echo "Usage:  "
    echo "  prepare_pwk.sh node_cnt(1/4/7/10/13/16) chain_cnt(1-4) p2p_port(default:11301) rpc_port(default:12301)"
    echo "    eg1: prepare_pwk.sh 4 1"
    echo "    eg2: prepare_pwk.sh 4 1 11301 12301"
}

if ( [ $# -eq 1 ] && [ "$1" ==  "-h" ] ) ; then
    show_help
    exit 1
fi

if [ ! $# -eq 2 ] && [ ! $# -eq 3 ] && [ ! $# -eq 4 ]; then
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
    echo "begin check params..."
    if  [[ ! -n $NODE_CNT ]] ;then
        echo "node cnt is empty"
        show_help
        exit 1
    fi

    if  [ ! $NODE_CNT -eq 1 ] && [ ! $NODE_CNT -eq 4 ] && [ ! $NODE_CNT -eq 7 ]&& [ ! $NODE_CNT -eq 10 ]&& [ ! $NODE_CNT -eq 13 ]&& [ ! $NODE_CNT -eq 16 ];then
        echo "node cnt should be 1 or 4 or 7 or 10 or 13 or 16"
        show_help
        exit 1
    fi

    if  [[ ! -n $CHAIN_CNT ]] ;then
        echo "chain cnt is empty"
        show_help
        exit 1
    fi

    if  [ ${CHAIN_CNT} -lt 1 ] || [ ${CHAIN_CNT} -gt 4 ] ;then
        echo "chain cnt should be 1 - 4"
        show_help
        exit 1
    fi

    if  [[ ! -n $P2P_PORT ]] ;then
        P2P_PORT=11301
    fi

    if  [ ${P2P_PORT} -ge 60000 ] || [ ${P2P_PORT} -le 10000 ];then
        echo "p2p_port_prefix should >=10000 && <=60000"
        show_help
        exit 1
    fi

    if  [[ ! -n $RPC_PORT ]] ;then
        RPC_PORT=12301
    fi

    if  [ ${RPC_PORT} -ge 60000 ] || [ ${RPC_PORT} -le 10000 ];then
        echo "rpc_port_prefix should >=10000 && <=60000"
        show_help
        exit 1
    fi
}

function generate_keys() {
    echo "begin generate keys, cnt: ${NODE_CNT}"
    mkdir -p ${BUILD_PATH}
    cd "${BUILD_PATH}"
    if [ -d crypto-config ]; then
        mkdir -p backup/backup_keys
        mv crypto-config  backup/backup_keys/crypto-config_$(date "+%Y%m%d%H%M%S")
    fi

    cp $CRYPTOGEN_TOOL_CONF crypto_config.yml
#    cp $CRYPTOGEN_TOOL_PKCS11_KEYS pkcs11_keys.yml

    xsed "s%count: 4%count: ${NODE_CNT}%g" crypto_config.yml

    ${CRYPTOGEN_TOOL_BIN} generate-pwk -c ./crypto_config.yml #-p ./pkcs11_keys.yml
}

function generate_config() {
    LOG_LEVEL="INFO"
    CONSENSUS_TYPE=1
    HASH_TYPE="SHA256"
    MONITOR_PORT=14321
    PPROF_PORT=24321
    TRUSTED_PORT=13301
    DOCKER_VM_PORT=22351
    DOCKER_VM_CONTAINER_NAME_PREFIX="chainmaker-vm-docker-go-container"
    ENABLE_DOCKERVM="false"

    read -p "input consensus type (0-SOLO,1-TBFT(default),3-MAXBFT,4-RAFT): " tmp
    if  [ ! -z "$tmp" ] ;then
      if  [ $tmp -eq 0 ] || [ $tmp -eq 1 ] || [ $tmp -eq 3 ] || [ $tmp -eq 4 ] ;then
          CONSENSUS_TYPE=$tmp
      else
        echo "unknown consensus type [" $tmp "], so use default"
      fi
    fi

    read -p "input log level (DEBUG|INFO(default)|WARN|ERROR): " tmp
    if  [ ! -z "$tmp" ] ;then
      if  [ $tmp == "DEBUG" ] || [ $tmp == "INFO" ] || [ $tmp == "WARN" ] || [ $tmp == "ERROR" ];then
          LOG_LEVEL=$tmp
      else
        echo "unknown log level [" $tmp "], so use default"
      fi
    fi

    read -p "input hash type (SHA256(default)|SM3): " tmp
    if  [ ! -z "$tmp" ] ;then
      if  [ $tmp == "SHA256" ] || [ $tmp == "SM3" ] ;then
          HASH_TYPE=$tmp
      else
        echo "unknown hash type [" $tmp "], so use default"
      fi
    fi

    read -p "enable docker vm (YES|NO(default))" enable_dockervm
        if  [ ! -z "$enable_dockervm" ]; then
          if  [ $enable_dockervm == "YES" ]; then
              ENABLE_DOCKERVM="true"
              echo "enable docker vm"
          fi
        fi

    cd "${BUILD_PATH}"
    if [ -d config ]; then
        mkdir -p backup/backup_config
        mv config  backup/backup_config/config_$(date "+%Y%m%d%H%M%S")
    fi

    mkdir -p ${BUILD_PATH}/config
    cd ${BUILD_PATH}/config

    node_count=$(ls -l $BUILD_CRYPTO_CONFIG_PATH|grep "^d"| wc -l)
    echo "config node total $node_count"
    for ((i = 1; i < $node_count + 1; i = i + 1)); do
        echo "begin generate node$i config..."
        mkdir -p ${BUILD_PATH}/config/node$i
        mkdir -p ${BUILD_PATH}/config/node$i/chainconfig
        cp $CONFIG_TPL_PATH/log.tpl node$i/log.yml
        xsed "s%{log_level}%$LOG_LEVEL%g" node$i/log.yml
        cp $CONFIG_TPL_PATH/chainmaker.tpl node$i/chainmaker.yml

        xsed "s%{net_port}%$(($P2P_PORT+$i-1))%g" node$i/chainmaker.yml
        xsed "s%{rpc_port}%$(($RPC_PORT+$i-1))%g" node$i/chainmaker.yml
        xsed "s%{monitor_port}%$(($MONITOR_PORT+$i-1))%g" node$i/chainmaker.yml
        xsed "s%{pprof_port}%$(($PPROF_PORT+$i-1))%g" node$i/chainmaker.yml
        xsed "s%{trusted_port}%$(($TRUSTED_PORT+$i-1))%g" node$i/chainmaker.yml
        xsed "s%{enable_dockervm}%$ENABLE_DOCKERVM%g" node$i/chainmaker.yml
        xsed "s%{docker_vm_port}%"$(($DOCKER_VM_PORT+$i-1))"%g" node$i/chainmaker.yml

        system=$(uname)

        if [ "${system}" = "Linux" ]; then
            for ((k = $NODE_CNT; k > 0; k = k - 1)); do
                xsed "/  seeds:/a\    - \"/ip4/127.0.0.1/tcp/$(($P2P_PORT+$k-1))/p2p/{org${k}_peerid}\"" node$i/chainmaker.yml
            done
        else
            ver=$(sw_vers | grep ProductVersion | cut -d':' -f2 | tr -d ' ')
            version=${ver:1:2}
            if [ $version -ge 11 ]; then
                for ((k = $NODE_CNT; k > 0; k = k - 1)); do
                xsed  "/  seeds:/a\\
    - \"/ip4/127.0.0.1/tcp/$(($P2P_PORT+$k-1))/p2p/{org${k}_peerid}\"\\
" node$i/chainmaker.yml
                done
            else
                for ((k = $NODE_CNT; k > 0; k = k - 1)); do
                  xsed  "/  seeds:/a\\
                  \ \ \ \ - \"/ip4/127.0.0.1/tcp/$(($P2P_PORT+$k-1))/p2p/{org${k}_peerid}\"\\
                  " node$i/chainmaker.yml
                done
            fi
        fi

        for ((j = 1; j < $CHAIN_CNT + 1; j = j + 1)); do
            xsed "s%#\(.*\)- chainId: chain${j}%\1- chainId: chain${j}%g" node$i/chainmaker.yml
            xsed "s%#\(.*\)genesis: ../config/{org_path${j}}/chainconfig/bc${j}.yml%\1genesis: ../config/{org_path${j}}/chainconfig/bc${j}.yml%g" node$i/chainmaker.yml

            if  [ $NODE_CNT -eq 1 ]; then
                if [ $CONSENSUS_TYPE -eq 0 ]; then
                    cp $CONFIG_TPL_PATH/chainconfig/bc_solo.tpl node$i/chainconfig/bc$j.yml
                    xsed "s%{consensus_type}%0%g" node$i/chainconfig/bc$j.yml
                else
                    cp $CONFIG_TPL_PATH/chainconfig/bc_solo.tpl node$i/chainconfig/bc$j.yml
                    xsed "s%{consensus_type}%$CONSENSUS_TYPE%g" node$i/chainconfig/bc$j.yml
                fi
            elif [ $NODE_CNT -eq 4 ] || [ $NODE_CNT -eq 7 ]; then
                cp $CONFIG_TPL_PATH/chainconfig/bc_4_7.tpl node$i/chainconfig/bc$j.yml
                xsed "s%{consensus_type}%$CONSENSUS_TYPE%g" node$i/chainconfig/bc$j.yml
            elif [ $NODE_CNT -eq 16 ]; then
                cp $CONFIG_TPL_PATH/chainconfig/bc_16.tpl node$i/chainconfig/bc$j.yml
                xsed "s%{consensus_type}%$CONSENSUS_TYPE%g" node$i/chainconfig/bc$j.yml
            else
                cp $CONFIG_TPL_PATH/chainconfig/bc_10_13.tpl node$i/chainconfig/bc$j.yml
                xsed "s%{consensus_type}%$CONSENSUS_TYPE%g" node$i/chainconfig/bc$j.yml
            fi

            xsed "s%{chain_id}%chain$j%g" node$i/chainconfig/bc$j.yml
            xsed "s%{hash_type}%$HASH_TYPE%g" node$i/chainconfig/bc$j.yml
            xsed "s%{version}%$VERSION%g" node$i/chainconfig/bc$j.yml

            if  [ $NODE_CNT -eq 7 ] || [ $NODE_CNT -eq 13 ] || [ $NODE_CNT -eq 16 ]; then
                xsed "s%#\(.*\)- org_id:%\1- org_id:%g" node$i/chainconfig/bc$j.yml
                xsed "s%#\(.*\)node_id:%\1node_id:%g" node$i/chainconfig/bc$j.yml
                xsed "s%#\(.*\)address:%\1address:%g" node$i/chainconfig/bc$j.yml
                xsed "s%#\(.*\)root:%\1root:%g" node$i/chainconfig/bc$j.yml
                xsed "s%#\(.*\)- \"%\1- \"%g" node$i/chainconfig/bc$j.yml
            fi

            if [ $NODE_CNT -eq 4 ] || [ $NODE_CNT -eq 7 ]; then
              xsed "${BC_YML_TRUST_ROOT_LINE},${BC_YML_TRUST_ROOT_LINE_END}d" node$i/chainconfig/bc$j.yml
            fi
            echo "begin node$i chain$CHAIN_CNT key config..."

            c=0
            for file in `ls -tr $BUILD_CRYPTO_CONFIG_PATH`
            do
                c=$(($c+1))
                xsed "s%{org${c}_id}%$file%g" node$i/chainconfig/bc$j.yml

                peerId=`cat $BUILD_CRYPTO_CONFIG_PATH/$file/node/consensus1/consensus1.nodeid`
                xsed "s%{org${c}_peerid}%$peerId%g" node$i/chainconfig/bc$j.yml

                if  [ $j -eq 1 ]; then
                    xsed "s%{org${c}_peerid}%$peerId%g" node$i/chainmaker.yml
                fi

                #cp admin
                mkdir -p $BUILD_CONFIG_PATH/node$i/keys/admin/$file
                cp $BUILD_CRYPTO_CONFIG_PATH/$file/admin/admin.pem $BUILD_CONFIG_PATH/node$i/keys/admin/$file

                if  [ $c -eq $i ]; then
                    if [ $c -gt $NODE_CNT ]; then
                      xsed "s%{node_pk_path}%node\/common1\/common1%g" node$i/chainmaker.yml
                      xsed "s%{net_pk_path}%node\/common1\/common1%g" node$i/chainmaker.yml
                    else
                      xsed "s%{node_pk_path}%node\/consensus1\/consensus1%g" node$i/chainmaker.yml
                      xsed "s%{net_pk_path}%node\/consensus1\/consensus1%g" node$i/chainmaker.yml
                    fi
                    xsed "s%{org_path}%$file%g" node$i/chainconfig/bc$j.yml
                    xsed "s%{node_pk_path}%node\/consensus1\/consensus1%g" node$i/chainmaker.yml
                    xsed "s%{net_pk_path}%node\/consensus1\/consensus1%g" node$i/chainmaker.yml
                    xsed "s%{org_id}%$file%g" node$i/chainmaker.yml
                    xsed "s%{org_path}%$file%g" node$i/chainmaker.yml
                    xsed "s%{org_path$j}%$file%g" node$i/chainmaker.yml

                    cp -r $BUILD_CRYPTO_CONFIG_PATH/$file/node $BUILD_CONFIG_PATH/node$i/keys
                    cp -r $BUILD_CRYPTO_CONFIG_PATH/$file/user $BUILD_CONFIG_PATH/node$i/keys
                fi

            done
        done

        echo "begin node$i trust config..."
        if  [ $NODE_CNT -eq 4 ] || [ $NODE_CNT -eq 7 ]; then
          for ((k = 1; k < $CHAIN_CNT + 1; k = k + 1)); do
            for file in `ls -t $BUILD_CRYPTO_CONFIG_PATH`
            do
                org_id_tmp="\ - org_id: \"${file}\""
                org_root="\ \ \ root:"
                org_root_tmp="\ \ \ \ \ - \"../config/wx-org${i}.chainmaker.org/keys/admin/${file}/admin.pem\""
                if [ "${system}" = "Linux" ]; then
                  xsed "${BC_YML_TRUST_ROOT_LINE}i\ ${org_root_tmp}" node$i/chainconfig/bc$k.yml
                  xsed "${BC_YML_TRUST_ROOT_LINE}i\ ${org_root}" node$i/chainconfig/bc$k.yml
                  xsed "${BC_YML_TRUST_ROOT_LINE}i\ ${org_id_tmp}"   node$i/chainconfig/bc$k.yml
                else
                  xsed "${BC_YML_TRUST_ROOT_LINE}i\\
\ ${org_root_tmp}\\
" node$i/chainconfig/bc$k.yml
                  xsed "${BC_YML_TRUST_ROOT_LINE}i\\
\ ${org_root}\\
"  node$i/chainconfig/bc$k.yml
                  xsed "${BC_YML_TRUST_ROOT_LINE}i\\
\ ${org_id_tmp}\\
"    node$i/chainconfig/bc$k.yml
                fi
            done
          done
        fi
    done
}

check_params
generate_keys
generate_config
