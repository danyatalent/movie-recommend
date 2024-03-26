#!/bin/bash

# Параметры по умолчанию
REMOTE_USER=""
REMOTE_HOST=""
REMOTE_PORT="22"
REMOTE_PATH="~/deploy/bin"
LOCAL_APP_PATH="./build/app"
CONFIG_FILE=""
REMOTE_CONFIG_PATH="~/configs/"

# Функция для вывода справки
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -u, --username      Username for SSH connection"
    echo "  -h, --host          Hostname or IP address of the remote server"
    echo "  -p, --port          SSH port (default: 22)"
    echo "  -r, --remote-path   Remote path to deploy the application"
    echo "  -l, --local-path    Local path to your web application"
    echo "  -c, --config        Path to the configuration file"
    echo "  -help               Display this help and exit"
}

# Обработка аргументов командной строки
while getopts ":u:h:p:r:l:c:" opt; do
    case ${opt} in
        u | --username )
            REMOTE_USER=$OPTARG
            ;;
        h | --host )
            REMOTE_HOST=$OPTARG
            ;;
        p | --port )
            REMOTE_PORT=$OPTARG
            ;;
        r | --remote-path )
            REMOTE_PATH=$OPTARG
            ;;
        l | --local-path )
            LOCAL_APP_PATH=$OPTARG
            ;;
        c | --config )
            CONFIG_FILE=$OPTARG
            ;;
        \? | : | * )
            usage
            exit 1
            ;;
    esac
done
shift $((OPTIND -1))

# Проверка обязательных параметров
if [ -z "$REMOTE_USER" ] || [ -z "$REMOTE_HOST" ]; then
    echo "Error: Username and host are required."
    usage
    exit 1
fi

# Функция для сборки приложения
build_app() {
    go build -o build/app github.com/danyatalent/movie-recommend/cmd/main
}

# Функция для копирования собранного приложения на удаленный сервер
deploy_app() {
    scp -P $REMOTE_PORT $LOCAL_APP_PATH $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH
}


# Сборка приложения
build_app

# Доставка на удаленный сервер
deploy_app


echo "Deployment completed successfully."
