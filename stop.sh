kn service delete --all
./scripts/github_runner/clean_cri_runner.sh
rm -f vhive
rm -f /tmp/vhive-logs/*
sudo rm -f /run/containerd/s/*
sudo ../dsa-perf-micros/scripts/setup_dsa.sh -d dsa0
