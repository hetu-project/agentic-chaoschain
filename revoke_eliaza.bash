#!/bin/bash


# 加载 nvm 环境变量
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # 加载 nvm




# 使用 nvm 切换到 Node.js 23 版本
nvm use 23 || { echo "无法切换到 Node.js 23 版本，请确保 nvm 已安装且版本存在。"; exit 1; }



for one in "alice" "bob" "charlie"; do
    if [ -f "out$i.log" ]; then
        rm -rf out$i.log
    fi
done


num=26657
i=1
for one in "alice" "bob" "charlie"; do
# for one in "alice"; do
    data_dir="data$i"
    if [ -f "agent/$one.json" ]; then
        cp "agent/$one.json" "agent/$one.json.cp" 
        mv "agent/$one.json" "agent/eliza.json"
        mv "agent/$one.json.cp" "agent/$one.json"
        echo "已将 agent/$one.json 重命名为 agent/eliza.json"
    else
        echo "文件 agent/$one.json 不存在，跳过。"
    fi

    # 启动 pnpm 服务
    sed -i "s#^SQLITE_FILE=.*#SQLITE_FILE=\"/root/hetu-chaoschain/hac-node/build/$data_dir/data/eliza.db\"#" .env
    sed -i "s#^COMET_URL=.*#COMET_URL=\"http://127.0.0.1:$num\"#" .env
    sed -i "s#^COMET_PRIVKEY=.*#COMET_PRIVKEY=\"/root/hetu-chaoschain/hac-node/build/$data_dir/config/priv_validator_key.json\"#" .env

    nohup pnpm start > out${one}.log 2>&1 &

    echo "服务已启动，日志输出到 out$one"
    ((num+=10))  # num 每次加 10
    ((i++)) 
done

