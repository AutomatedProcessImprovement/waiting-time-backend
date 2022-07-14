#!/usr/bin/env bash

remote_host=193.40.11.233
deployments_dir=/home/ihar/deployments
destination_dir=$deployments_dir/waiting-time-backend
payload_dir=build/linux-amd64

#rsync -avz $payload_dir/ $remote_host:$destination_dir/
#
#rsync -avz $service_file $remote_host:$destination_dir/
#ssh $remote_host sudo -S mv $destination_dir/$service_file /etc/systemd/system/
#
#ssh $remote_host sudo -S systemctl daemon-reload
#ssh $remote_host sudo -S systemctl restart $service_file

rsync -avz caddy $remote_host:$deployments_dir/
ssh $remote_host chmod +x $deployments_dir/caddy/caddy_linux_amd64

# TODO: finish publishing the web service and exposing it to the internet