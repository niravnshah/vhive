FROM frolvlad/alpine-python3 as as builder_workload
RUN apk add --no-cache \
    g++ gcc gfortran file binutils linux-headers \
    musl-dev python3-dev cython openblas-dev lapack-dev \
    libstdc++ openblas lapack py3-scipy
COPY requirements.txt .
RUN pip3 install --no-cache-dir -r requirements.txt

FROM ghcr.io/niravnshah/vhive/lr_serving:builder_workload as var_workload
COPY *.py /
COPY dataset.csv /
COPY lr_model.pk /

EXPOSE 50051

STOPSIGNAL SIGKILL

CMD ["/usr/bin/python", "/server.py"]
