#!/bin/sh

while true; do cp /data/taskmanager.rdb /backup/$(date +%s ).rdb; sleep $BACKUP_PERIOD; done 