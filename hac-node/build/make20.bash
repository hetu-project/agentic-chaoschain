#!/bin/bash
ps -ef | grep './hac -d' | awk '{print $2}' | xargs kill -9
sleep 1

NUM_NODES=20
BASE_P2P_PORT=26656
BASE_RPC_PORT=26657
BASE_APP_PORT=26658
PORT_INTERVAL=10


for i in $(seq 1 $((NUM_NODES))); do
    if [ -d "data$i" ]; then
        rm -rf data$i
    fi
done


for i in $(seq 1 $((NUM_NODES))); do
    ./hac init --home data$i
    sleep 2
done

VALIDATORS_JSON="[]"
AGENTS_JSON="[]"
for i in $(seq 1 $((NUM_NODES))); do
    sleep 1
    VALIDATOR=$(cat data$i/config/genesis.json | jq '.validators[0]')
    AGENT=$(cat data$i/config/genesis.json | jq '.app_state.agents[0]')
    VALIDATORS_JSON=$(echo $VALIDATORS_JSON | jq ". + [$VALIDATOR]")
    AGENTS_JSON=$(echo $AGENTS_JSON | jq ". + [$AGENT]")
    echo "1111111"

    echo $VALIDATOR
    echo $AGENT
    echo "1111111"

done
echo $VALIDATORS_JSON
echo $AGENTS_JSON

cat data1/config/genesis.json | \
  jq ".validators = $VALIDATORS_JSON | .app_state.agents = $AGENTS_JSON" \
  > data1/config/genesis.json.tmp && mv data1/config/genesis.json.tmp data1/config/genesis.json

for i in $(seq 2 $((NUM_NODES))); do
    cp data1/config/genesis.json data$i/config/
done

declare -a NODE_IDS
for i in $(seq 1 $((NUM_NODES))); do
    NODE_IDS[$i]=$(cometbft show-node-id --home data$i)
    echo ${NODE_IDS[$i]}
done

for i in $(seq 1 $((NUM_NODES))); do
    sleep 1
    P2P_PORT=$((BASE_P2P_PORT + (i-1)*PORT_INTERVAL))
    RPC_PORT=$((BASE_RPC_PORT + (i-1)*PORT_INTERVAL))
    APP_PORT=$((BASE_APP_PORT + (i-1)*PORT_INTERVAL))
    
    #sed -i "s/^proxy_app = \"tcp:\/\/127.0.0.1:26658\"/proxy_app = \"tcp:\/\/127.0.0.1:$APP_PORT\"/" data$i/config/config.toml
    sed -i "s/^laddr = \"tcp:\/\/0.0.0.0:26656\"/laddr = \"tcp:\/\/0.0.0.0:$P2P_PORT\"/" data$i/config/config.toml
    sed -i "s/^laddr = \"tcp:\/\/127.0.0.1:26657\"/laddr = \"tcp:\/\/127.0.0.1:$RPC_PORT\"/" data$i/config/config.toml
    sed -i "s/addr_book_strict = true/addr_book_strict = false/" data$i/config/config.toml
    sed -i "s/pex = true/pex = false/" data$i/config/config.toml


    PEERS=""
    for j in $(seq 1 $((NUM_NODES))); do
        PEER_PORT=$((BASE_P2P_PORT + (j-1)*PORT_INTERVAL))
        if [ -z "$PEERS" ]; then
            PEERS="${NODE_IDS[$j]}@198.13.45.187:$PEER_PORT"
        else
            PEERS="${PEERS},${NODE_IDS[$j]}@198.13.45.187:$PEER_PORT"
        fi
    done
    

    sed -i "s/^persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" data$i/config/config.toml
    sed -i "s/^allow_duplicate_ip = false/allow_duplicate_ip = true/" data$i/config/config.toml

    SERVICE_PORT=$((8630 + i))
    sed -i "/^\[app\]/a\\
agent_url = \"http://127.0.0.1:3000\" # eliza agent service address\\
service_address = \"0.0.0.0:${SERVICE_PORT}\" # api server listen address\\
discussion_rate = 2 # controls the rate of discussion" data$i/config/config.toml
done


for i in $(seq 1 $((NUM_NODES))); do
    nohup ./hac -d data$i > out$i 2>&1 &
    sleep 0.2
done