#!/bin/bash
set -e

SHARE_DIST_DIR=${SHARE_DIST_DIR:-"/shared-dir/dist/plugins/"}
SHARE_PLUGINS_DIR=${SHARE_PLUGINS_DIR:-"/shared-dir/plugins/"}

WORK_DIR=${WORK_DIR:-"/app"}

CONFIG_DIR=$WORK_DIR/config
CONFIG_FILE=${CONFIG_FILE:-"$CONFIG_DIR/config.toml"}
RUN_CONFIG_FILE=${RUN_CONFIG_FILE:-"$CONFIG_DIR/testrunner-deployment-template.yaml"}
STATIC_ROOT_DIR=$WORK_DIR/front

mkdir -p $WORK_DIR $CONFIG_DIR $STATIC_ROOT_DIR $SHARE_DIST_DIR $SHARE_PLUGINS_DIR

EXEC_BIN=$WORK_DIR/backend

# Define the service start command
SERVICE_COMMAND="$EXEC_BIN run --config $CONFIG_FILE --static-root $STATIC_ROOT_DIR --runner-cfg $RUN_CONFIG_FILE"

# Define the cleanup function
cleanup() {
    echo "Stopping service..."
    kill $SERVICE_PID
    wait $SERVICE_PID
    echo "Service stopped."
}

# Capture signals
trap cleanup SIGINT SIGTERM

# start docker service
echo "Starting docker service..."
nohup /usr/bin/dockerd >/var/log/dockerd.log 2>&1 &

# Start the service and put it in the background
echo "Starting service..."
$SERVICE_COMMAND &
SERVICE_PID=$!
echo "Service started with PID: $SERVICE_PID"

# # do something for init
# RETRY_INTERVAL=3
# retry_count=0
# LABEL="import plugin [$SHARE_DIST_DIR]"
# set +e
# while true; do
#     $EXEC_BIN import-plugin $SHARE_DIST_DIR

#     if [ $? -eq 0 ]; then
#         echo "import plugin [$SHARE_DIST_DIR] executed successfully."
#         break # 命令成功执行，退出循环
#     else
#         retry_count=$((retry_count + 1))
#         echo "$LABEL $retry_count st failed"
#         set -e && sleep $RETRY_INTERVAL && set +e
#     fi
#     if [ $retry_count -ge 100 ]; then
#         echo "$LABEL failed after $retry_count attempts. Giving up."
#         break
#     fi
# done

set -e
# Keep the container running until the service ends
echo "Waiting for service to exit..."
wait $SERVICE_PID

echo "Service exited."
