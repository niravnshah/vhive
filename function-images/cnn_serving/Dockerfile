FROM tatsushid/alpine-py3-tensorflow-jupyter as builder_workload
COPY requirements.txt .
ENV GRPC_PYTHON_VERSION 1.26.0
COPY *.apk /
RUN apk add --no-cache --allow-untrusted /alpine-keys-2.4-r0.apk
RUN apk update && \
    apk add python3 python3-dev py3-pip && \
    ln -s /usr/bin/pip3 /usr/bin/pip && \
    ln -sf /usr/bin/pip3 /usr/local/bin/pip && \
    ln -sf /usr/bin/python3 /usr/local/bin/python && \
    ln -sf /usr/bin/python3 /usr/local/bin/python3 && \
    ln -s /usr/include/locale.h /usr/include/xlocale.h
RUN apk add --upgrade
RUN apk add --update --no-cache build-base gcc g++ protobuf
RUN apk add --no-cache --allow-untrusted /libsz-1.0.6-r1.apk
RUN apk add --no-cache --allow-untrusted /hdf5-1.12.2-r1.apk
RUN apk add --no-cache --allow-untrusted /hdf5-cpp-1.12.2-r1.apk
RUN apk add --no-cache --allow-untrusted /hdf5-hl-1.12.2-r1.apk
RUN apk add --no-cache --allow-untrusted /hdf5-hl-cpp-1.12.2-r1.apk
RUN apk del openblas py3-numpy py3-numpy-f2py libgfortran
RUN apk add --no-cache --allow-untrusted /libgfortran-10.3.1_git20210424-r2.apk
RUN apk add --no-cache --allow-untrusted /hdf5-fortran-1.12.2-r1.apk
RUN apk add --no-cache --allow-untrusted /hdf5-hl-fortran-1.12.2-r1.apk
RUN apk add --no-cache --allow-untrusted /hdf5-dev-1.12.2-r1.apk
RUN apk add --no-cache --allow-untrusted /openblas-0.3.15-r0.apk
RUN apk add --no-cache --allow-untrusted /py3-numpy-1.20.3-r0.apk
RUN apk add --no-cache --allow-untrusted /py3-numpy-dev-1.20.3-r0.apk
RUN apk add --no-cache --allow-untrusted /py3-numpy-f2py-1.20.3-r0.apk
RUN pip3 install --upgrade pip
RUN pip3 install --no-cache-dir Cython
RUN pip3 uninstall -y enum34
RUN pip3 install --no-cache-dir -r requirements.txt
RUN pip3 install --no-cache-dir protobuf==3.11.3 grpcio==${GRPC_PYTHON_VERSION}

FROM tatsushid/alpine-py3-tensorflow-jupyter as var_workload
COPY *.py /
COPY image* /
COPY squeezenet_weights_tf_dim_ordering_tf_kernels.h5  /tmp

EXPOSE 50051

STOPSIGNAL SIGKILL

CMD ["/usr/bin/python", "/server.py"]
