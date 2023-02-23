FROM ghcr.io/niravnshah/vhive/py_grpc:builder_grpc as builder_workload
COPY --from=ghcr.io/niravnshah/vhive/py_grpc:builder_grpc /root/.local /root/.local
COPY requirements.txt .
RUN pip3 install --user -r requirements.txt

FROM ghcr.io/niravnshah/vhive/py_grpc:base as var_workload
ENV PYTHONUNBUFFERED=1
COPY *.py /
COPY --from=builder_workload /root/.local /root/.local
RUN apk add libstdc++ --update --no-cache

EXPOSE 50051

STOPSIGNAL SIGKILL

CMD ["/usr/local/bin/python", "/server.py"]
