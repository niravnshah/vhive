FROM ghcr.io/niravnshah/vhive/py_grpc:builder_grpc as builder_workload
COPY --from=ghcr.io/niravnshah/vhive/py_grpc:builder_grpc /root/.local /root/.local
COPY requirements.txt .
RUN apk add --no-cache --virtual .build-deps \
    build-base linux-headers zlib-dev jpeg-dev \
    && pip3 install --user -r requirements.txt \
    && apk del .build-deps

FROM ghcr.io/niravnshah/vhive/py_grpc:base as var_workload
COPY *.py /
COPY *.jpeg /
COPY --from=builder_workload /root/.local /root/.local
RUN apk add libstdc++ libjpeg-turbo --update --no-cache

EXPOSE 50051

STOPSIGNAL SIGKILL

CMD ["/usr/local/bin/python", "/server.py"]
