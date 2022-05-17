#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# [*] the represented items could not be modified after startup

# "auth_type" should be consistent among the whole chain configuration files(e.g., bc1.yml and chainmaker.yml)
# The auth type can be permissionedWithCert, permissionedWithKey, public.
# By default it is permissionedWithCert.
# permissionedWithCert: permissioned blockchain, using x.509 certificate to identify members.
# permissionedWithKey: permissioned blockchain, using public key to identify members.
# public: public blockchain, using public key to identify members.
auth_type: "permissionedWithKey" # [*]

# Logger settings
log:
  # Logger configuration file path.
  config_file: ../config/{org_path}/log.yml

# Crypto engine config
crypto_engine: tjfoc #support gmssl, tencentsm and tjfoc

# Chains the node currently joined in
blockchain:
  # chain id and its genesis block file path.
#  - chainId: chain1
#    genesis: ../config/{org_path1}/chainconfig/bc1.yml
#  - chainId: chain2
#    genesis: ../config/{org_path2}/chainconfig/bc2.yml
#  - chainId: chain3
#    genesis: ../config/{org_path3}/chainconfig/bc3.yml
#  - chainId: chain4
#    genesis: ../config/{org_path4}/chainconfig/bc4.yml


# Blockchain node settings
node:
  # Organization id is the node belongs to.
  # When the auth type is public, org id is ignored.
  org_id:            {org_id}  # [*]

  # Private key file path
  priv_key_file: ../config/{org_path}/keys/{node_pk_path}.key # [*]

  # Certificate cache size, used to speed up member identity verification.
  # By default the cache size is 1000.
  cert_cache_size:   1000

  # fast sync settings
  fast_sync:
    # Enable it or not
    enabled: false  # [*]

    # The number of blocks that did not perform fast synchronization at the end
    min_full_blocks: 10

  # PKCS#11 crypto settings
  pkcs11:
    # Enable it or not
    enabled: false  # [*]

    # Path for the pkcs11 interface file(.so)
    library: /usr/local/lib64/pkcs11/libupkcs11.so

    # Label for the slot to be used
    label: HSM

    # HSM Password
    password: 11111111

    # Size for HSM session cache, default value is 10.
    session_cache_size: 10

    # Hash algorithm is used to calculate SKI.
    # It can be SHA256 or SM3.
    hash: "SHA256"  # [*]

# Network Settings
net:
  # Network provider, can be libp2p or liquid.
  # libp2p: using libp2p components to build the p2p module.
  # liquid: a new p2p network module. We build it from 0 to 1.
  # This item must be consistent across the blockchain network.
  provider: LibP2P

  # The address and port the node listens on.
  # By default, it uses 0.0.0.0 to listen on all network interfaces.
  listen_addr: /ip4/0.0.0.0/tcp/{net_port}

  # Max stream of a connection.
  # peer_stream_pool_size: 100

  # Max number of peers the node can connect.
  # max_peer_count_allow: 20

  # The strategy for eliminating node when the amount of connected peers reaches the max value
  # It could be: 1 Random, 2 FIFO, 3 LIFO. The default strategy is LIFO.
  # peer_elimination_strategy: 3

  # The seeds list used to setup network among all the peer seed when system starting.
  # The connection supervisor will try to dial seed peer whenever the connection is broken.
  # Example ip format: "/ip4/127.0.0.1/tcp/11301/p2p/"+nodeid
  # Example dns format："/dns/cm-node1.org/tcp/11301/p2p/"+nodeid
  seeds:

  # Network tls settings.
  tls:
    # Enable tls or not. Currently it can only be true...
    enabled: true

    # TLS private key file path.
    priv_key_file: ../config/{org_path}/keys/{net_pk_path}.key

  # The blacklist is automatically block the listed seed to connect.
  # blacklist:
      # The addresses in blacklist.
      # The address format can be ip or ip+port.
      # addresses:
      #   - "127.0.0.1:11301"
      #   - "192.168.1.8"

      # The node ids in blacklist.
      # node_ids:
      #   - "QmeyNRs2DwWjcHTpcVHoUSaDAAif4VQZ2wQDQAUNDP33gH"

# Transaction pool settings
# Other txpool settings can be found in tx_Pool_config.go
txpool:
  # txpool type, can be signle or batch.
  # By default the txpool type is single.
  pool_type: "single"

  # Max transaction count in txpool.
  # If txpool is full, the following transactions will be discarded.
  max_txpool_size: 50000

  # Max config transaction count in config txpool.
  max_config_txpool_size: 10

  # Interval of creating a transaction batch, only for batch txpool, in millisecond.
  # batch_create_timeout: 200


# RPC service setting
rpc:
  # RPC type, can only be grpc now
  provider: grpc  # [*]

  # RPC port
  port: {rpc_port}

  # Interval of checking trust root changes, in seconds.
  # If changed, the rpc server's root certificate pool will also change.
  # Only valid if tls is enabled.
  # The minium value is 10.
  check_chain_conf_trust_roots_change_interval: 60

  # Rate limit related settings
  # Here we use token bucket to limit rate.
  ratelimit:
    # Ratelimit switch. Default is false.
    enabled: false

    # Rate limit type
    # 0: limit globally, 1: limit by ip
    type: 0

    # Token number added to bucket per second.
    # -1: unlimited, by default is 10000.
    token_per_second: -1

    # Token bucket size.
    # -1: unlimited, by default is 10000.
    token_bucket_size: -1

  # Rate limit settings for subscriber
  subscriber:
    ratelimit:
      token_per_second: 100
      token_bucket_size: 100

  # RPC TLS settings
  tls:
    # TLS mode, can be disable, oneway, twoway.
    mode: disable

  # RPC blacklisted ip addresses
  blacklist:
    addresses:
      # - "127.0.0.1"

  # RPC server max send/receive message size in MB
  max_send_msg_size: 10
  max_recv_msg_size: 10

# Transaction filter settings
tx_filter:
  # default(store) 0; bird's nest 1; map 2; 3 sharding bird's nest
  # 3 is recommended.
  type: 0
  # sharding bird's nest config
  # total keys = sharding.length * sharding.birds_nest.length * sharding.birds_nest.cuckoo.max_num_keys
  sharding:
    # sharding size
    length: 5
    # sharding task timeout in seconds
    timeout: 3
    snapshot:
      # serialize type
      # 0 Serialization by height interval
      # 1 Serialization by time interval
      type: 0
      block_height:
        # Block height interval
        interval: 10
      timed:
        # Time interval in seconds
        interval: 10
      # file path
      path: ../data/{org_id}/tx_filter
    # bird's nest config
    birds_nest:
      # bird's nest size
      length: 10
      # Transaction filter rules
      rules:
        # Absolute expiration time /second
        # Based on the number of transactions per day, for example, the current total capacity of blockchain transaction
        # filters is 100 million, and there are 10 million transaction requests per day.
        #
        # total keys = sharding.length * sharding.birds_nest.length * sharding.birds_nest.cuckoo.max_num_keys
        #
        # absolute expire time = total keys / number of requests per day
        absolute_expire_time: 172800
      cuckoo:
        # num of tags for each bucket, which is b in paper. tag is fingerprint, which is f in paper.
        # If you are using a semi-sorted bucket, the default is 4
        # 2 is recommended.
        tags_per_bucket: 2
        # num of bits for each item, which is length of tag(fingerprint)
        # 11 is recommended.
        bits_per_item: 11
        # keys number
        max_num_keys: 2000000
        # 0 TableTypeSingle normal single table
        # 1 TableTypePacked packed table, use semi-sort to save 1 bit per item
        # 0 is recommended
        table_type: 0
  # bird's nest config
  # total keys = birds_nest.length * birds_nest.cuckoo.max_num_keys
  birds_nest:
    # bird's nest size
    length: 10
    snapshot:
      # serialize type
      # 0 Serialization by height interval
      # 1 Serialization by time interval
      type: 0
      block_height:
        # Block height interval
        interval: 10
      timed:
        # Time interval in seconds
        interval: 10
      # file path
      path: ../data/{org_id}/tx_filter
    # Transaction filter rules
    rules:
      # Absolute expiration time /second
      # Based on the number of transactions per day, for example, the current total capacity of blockchain transaction
      # filters is 100 million, and there are 10 million transaction requests per day.
      #
      # total keys = sharding.length * sharding.birds_nest.length * sharding.birds_nest.cuckoo.max_num_keys
      #
      # absolute expire time = total keys / number of requests per day
      absolute_expire_time: 172800
    cuckoo:
      # num of tags for each bucket, which is b in paper. tag is fingerprint, which is f in paper.
      # If you are using a semi-sorted bucket, the default is 4
      # 2 is recommended.
      tags_per_bucket: 2
      # num of bits for each item, which is length of tag(fingerprint)
      # 11 is recommended.
      bits_per_item: 11
      # keys number
      max_num_keys: 2000000
      # 0 TableTypeSingle normal single table
      # 1 TableTypePacked packed table, use semi-sort to save 1 bit per item
      # 0 is recommended
      table_type: 0

# Monitor related settings
monitor:
  # Monitor service switch, default is false.
  enabled: false

  # Monitor service port
  port: {monitor_port}

# PProf Settings
pprof:
  # If pprof is enabled or not
  enabled: false

  # PProf port
  port: {pprof_port}

# Consensus related settings
consensus:
  raft:
    # Take a snapshot based on the set the number of blocks.
    # If raft nodes change, a snapshot is taken immediately.
    snap_count: 10

    # Saving wal asynchronously switch. Default is true.
    async_wal_save: true

    # Min time unit in rate election and heartbeat.
    ticker: 1

# Scheduler related settings
scheduler:
  # whether log the txRWSet map in debug mode
  rwset_log: false

# Storage config settings
# Contains blockDb, stateDb, historyDb, resultDb, contractEventDb
#
# blockDb: block transaction data,                          support leveldb, mysql, badgerdb, tikvdb
# stateDb: world state data,                                support leveldb, mysql, badgerdb, tikvdb
# historyDb: world state change history of transactions,    support leveldb, mysql, badgerdb, tikvdb
# resultDb: transaction execution results data,             support leveldb, mysql, badgerdb, tikvdb
# contractEventDb: contract emit event data,                support mysql
#
# provider, sqldb_type cannot be changed after startup.
# store_path, dsn the content cannot be changed after startup.
storage:
  # Default store path
  store_path: ../data/{org_id}/ledgerData1 # [*]

  # Prefix for mysql db name
  # db_prefix: org1_

  # Minimum block height not allowed to be archived
  unarchive_block_height: 300000

  # Symmetric encryption algorithm for writing data to disk. can be sm4 or aes
  # encryptor: sm4    # [*]

  # Disable block file db, default: false
  disable_block_file_db: false

  # async write block in file block db to disk, default: false, so default is sync write disk
  logdb_segment_async: false

  # file size of .fdb, MB, default: 20
  logdb_segment_size: 128

  # bigfilter config
  enable_bigfilter: false    #default false
  bigfilter_config:
    redis_hosts_port: "127.0.0.1:6300,127.0.0.1:6301"   #redis host:port
    redis_password: abcpass  #redis password
    tx_capacity: 1000000000   #support max transaction capacity
    fp_rate: 0.000000001      #false postive rate

  # RWC config
  enable_rwc: true   #default false

  # suggest
  # if block_tx_capacity < 10000,
  # set rolling_window_cache_capacity greater than block_tx_capacity*1.1 and less than block_tx_capacity*2
  # if block_tx_capacity > 10000,  set rolling_window_cache_capacity 20000
  rolling_window_cache_capacity: 200

  # Symmetric encryption key:16 bytes key
  # If pkcs11 is enabled, it is the keyID
  # encrypt_key: "1234567890123456"
  write_block_type: 0  # 0普通写模式，1快速写模式
  disable_state_cache: false # default false
  state_cache_config:
    life_window: 3000000000000   #key/value ttl 时间，单位 ns
    clean_window: 1000000000
    max_entry_size: 500
    hard_max_cache_size: 10240   #缓存大小，单位MB

  # Block db config
  blockdb_config:
    # Databases type support leveldb, sql, badgerdb, tikvdb
    provider: leveldb # [*]
    # If provider is leveldb, leveldb_config should not be null.
    leveldb_config:
      # LevelDb store path
      store_path: ../data/{org_id}/block

    # Example for sql provider
    # Databases type support leveldb, sql, badgerdb,tikvdb
    # provider: sql # [*]
    # If provider is sql, sqldb_config should not be null.
    # sqldb_config:
      # Sql db type, can be mysql, sqlite. sqlite only for test
      # sqldb_type: mysql # # [*]
      # Mysql connection info, the database name is not required. such as:  root:admin@tcp(127.0.0.1:3306)/
      # dsn: root:password@tcp(127.0.0.1:3306)/

    # Example for badgerdb provider
    # Databases type support leveldb, sql, badgerdb, tikvdb
    # provider: badgerdb
    # If provider is badgerdb, badgerdb_config should not be null.
    # badgerdb_config:
      # BadgerDb store path
      # store_path: ../data/wx-org1.chainmaker.org/history
      # Whether compression is enabled for stored data, default is 0: disabled
      # compression: 0
      # Key and value are stored separately when value is greater than this byte, default is 1024 * 10
      # value_threshold: 256
      # Number of key value pairs written in batch. default is 128
      # write_batch_size: 1024

    # Example for tikv provider
    # provider: tikvdb
    # If provider is tikvdb, tikvdb_config should not be null.
    # tikvdb_config:
      # db_prefix: "node1_" #default is ""
      # endpoints: "127.0.0.1:2379" # tikv pd server url，support multi url, example :"192.168.1.2:2379,192.168.1.3:2379"
      # max_batch_count: 128  # max tikv commit batch size, default: 128
      # grpc_connection_count: 16 # chainmaker and tikv connect count, default: 4
      # grpc_keep_alive_time: 10 # keep connnet alive count, default: 10
      # grpc_keep_alive_timeout: 3  # keep connnect alive time, default: 3
      # write_batch_size: 128 # commit tikv bacth size each time, default: 128
  # State db config
  statedb_config:
    provider: leveldb
    leveldb_config:
      store_path: ../data/{org_id}/state

  # History db config
  historydb_config:
    provider: leveldb
    leveldb_config:
      store_path: ../data/{org_id}/history

  # Result db config
  resultdb_config:
    provider: leveldb
    leveldb_config:
      store_path: ../data/{org_id}/result

  # Disable contract event database or not. If it is false, contract_eventdb_config must be mysql
  disable_contract_eventdb: true
  # Contract event db config
  contract_eventdb_config:
    # Event db only support sql
    provider: sql
    # Sql db config
    sqldb_config:
      # Event db only support mysql
      sqldb_type: mysql
      # Mysql connection info, such as:  root:admin@tcp(127.0.0.1:3306)/
      dsn: root:password@tcp(127.0.0.1:3306)/

# Docker go virtual machine configuration
vm:
  # Enable docker go virtual machine
  enable_dockervm: {enable_dockervm}
  # Mount point in chain maker
  dockervm_mount_path: ../data/{org_id}/docker-go
  # Specify log file path
  dockervm_log_path: ../log/{org_id}/docker-go
  # Whether to print log at terminal
  log_in_console: false
  # Log level
  log_level: INFO
  # Unix domain socket open, used for chainmaker and docker manager communication
  uds_open: true
  # docker vm contract service host, default 127.0.0.1
  docker_vm_host: 127.0.0.1
  # docker vm contract service port, default 22351
  docker_vm_port: {docker_vm_port}
  # Grpc max send message size, Default size is 4, Unit: MB
  max_send_msg_size: 20
  # Grpc max receive message size, Default size is 4, Unit: MB
  max_recv_msg_size: 20
  # max number of connection created to connect docker vm service
  max_connection: 5