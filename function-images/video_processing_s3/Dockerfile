FROM denismakogon/opencv3-slim:edge as builder_workload
RUN apt update && apt -y install build-essential libgl1-mesa-glx libglib2.0-0
RUN pip3 install --upgrade pip && pip3 install minio && pip3 install grpcio && pip3 install opencv-python && pip3 install protobuf

FROM ghcr.io/niravnshah/vhive/video_processing_s3:builder_workload as var_workload
ENV PYTHONUNBUFFERED=1
COPY *.py /
COPY *.mp4 /

EXPOSE 50051

STOPSIGNAL SIGKILL

CMD ["/usr/local/bin/python3", "/server.py"]
