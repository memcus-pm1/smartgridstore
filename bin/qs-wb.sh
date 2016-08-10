#!/bin/bash

ssh -t $SSH_USER@$TUNNEL_IP "sudo docker run -it -v /srv/sync:/etc/sync -v /srv/certs:/etc/letsencrypt -e CLUSTER_INFO=\"$CLUSTER_INFO\" immesys/btrdb-qs-wb"
