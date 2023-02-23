pushd grpc

docker pull ghcr.io/niravnshah/vhive/py_grpc:base
docker pull ghcr.io/niravnshah/vhive/py_grpc:builder_grpc

docker build --target base --cache-from=ghcr.io/niravnshah/vhive/py_grpc:base --tag ghcr.io/niravnshah/vhive/py_grpc:base . 
docker push ghcr.io/niravnshah/vhive/py_grpc:base
docker build --target builder_grpc --cache-from=ghcr.io/niravnshah/vhive/py_grpc:base --cache-from=ghcr.io/niravnshah/vhive/py_grpc:builder_grpc --tag ghcr.io/niravnshah/vhive/py_grpc:builder_grpc . 
docker push ghcr.io/niravnshah/vhive/py_grpc:builder_grpc
popd

for %%f in (chameleon cnn_serving image_rotate_s3 json_serdes_s3 lr_serving lr_training_s3 pyaes video_processing_s3 helloworld) do (
pushd %%f
docker pull ghcr.io/niravnshah/vhive/%%f:builder_workload
docker pull ghcr.io/niravnshah/vhive/%%f:var_workload

docker build --target builder_workload --cache-from=ghcr.io/niravnshah/vhive/py_grpc:base --cache-from=ghcr.io/niravnshah/vhive/py_grpc:builder_grpc --cache-from=ghcr.io/niravnshah/vhive/%%f:builder_workload --tag ghcr.io/niravnshah/vhive/%%f:builder_workload . 
docker push ghcr.io/niravnshah/vhive/%%f:builder_workload

docker build --target var_workload --cache-from=ghcr.io/niravnshah/vhive/py_grpc:base --cache-from=ghcr.io/niravnshah/vhive/py_grpc:builder_grpc --cache-from=ghcr.io/niravnshah/vhive/%%f:builder_workload --cache-from=ghcr.io/niravnshah/vhive/%%f:var_workload --tag ghcr.io/niravnshah/vhive/%%f:var_workload . 
docker push ghcr.io/niravnshah/vhive/%%f:var_workload
popd
)
