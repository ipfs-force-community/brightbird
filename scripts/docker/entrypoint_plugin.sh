#!/bin/bash

SHARE_DIST_DIR=${SHARE_DIST_DIR:-"/shared-dir/dist/"}
SHARE_PLUGINS_DIR=${SHARE_PLUGINS_DIR:-"/shared-dir/plugins/"}

mkdir -p $SHARE_DIST_DIR $SHARE_PLUGINS_DIR 
cp -r /plugins $SHARE_DIST_DIR