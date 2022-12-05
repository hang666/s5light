#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin
export PATH
LANG=en_US.UTF-8

VERSION="v0.1.1"
DEFAULT_BIND="0.0.0.0:8081"

BIN_DIR="/s5light"
CONFIG_PATH=$BIN_DIR/config.yaml


function install_s5light(){
    
    f_bindaddress=$1
    f_name=$2
    f_password=$3
    
    mkdir $BIN_DIR
    
    wget https://github.com/hang666/s5light/releases/download/${VERSION}/s5light-linux-amd64-${VERSION}.gz -O $BIN_DIR/s5light.gz
    gunzip -c $BIN_DIR/s5light.gz > $BIN_DIR/s5light
    rm -rf $BIN_DIR/s5light.gz
    chmod +x $BIN_DIR/s5light
    
    echo -e "accounts:\n  - username: \"$f_name\"\n    password: \"$f_password\"\n    bind_address: \"$f_bindaddress\"" > $CONFIG_PATH
    
    wget https://raw.githubusercontent.com/hang666/s5light/${VERSION}/config.yaml.example -O $BIN_DIR/config.yaml.example
    
    wget https://raw.githubusercontent.com/hang666/s5light/main/script/s5light.service -O /etc/systemd/system/s5light.service
    chmod +x /etc/systemd/system/s5light.service
    sudo systemctl daemon-reload
    sudo systemctl enable s5light
    sudo systemctl start s5light
    echo "Successfully started s5light!"
    echo "bind address:$f_bindaddress name:$f_name password:$f_password"
    echo "Configuration file path: ${CONFIG_PATH}"
    
}

echo "Please input listen address (0.0.0.0:8081): "
read bindaddress
echo "Please input auth name: "
read name
echo "Please input auth password: "
read password



if [ ! -n "$bindaddress" ]; then
    install_s5light $DEFAULT_BIND $name $password
else
    install_s5light $bindaddress $name $password
fi

