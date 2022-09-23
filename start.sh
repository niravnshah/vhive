sudo echo 2147483648 > /sys/bus/dsa/devices/wq0.0/max_transfer_size
sudo ../dsa-perf-micros/scripts/setup_dsa.sh -d dsa0 -w1 -e1 -ms
echo max_transfer_size = $(sudo cat /sys/bus/dsa/devices/wq0.0/max_transfer_size)
echo block_on_fault = $(sudo cat /sys/bus/dsa/devices/wq0.0/block_on_fault)
mkdir -p /tmp/vhive-logs
sudo screen -dmS containerd bash -c "containerd > >(tee -a /tmp/vhive-logs/containerd.stdout) 2> >(tee -a /tmp/vhive-logs/containerd.stderr >&2)"; sleep 5;
sudo PATH=$PATH screen -dmS firecracker bash -c "/usr/local/bin/firecracker-containerd --config /etc/firecracker-containerd/config.toml > >(tee -a /tmp/vhive-logs/firecracker.stdout) 2> >(tee -a /tmp/vhive-logs/firecracker.stderr >&2)"; sleep 5;
source /etc/profile && go build;
sudo screen -dmS vhive bash -c "./vhive -snapshots -upf -inmem -dsa > >(tee -a /tmp/vhive-logs/vhive.stdout) 2> >(tee -a /tmp/vhive-logs/vhive.stderr >&2)"; sleep 5;
sudo screen -ls
sudo kubeadm config images pull

./scripts/cluster/create_one_node_cluster.sh

