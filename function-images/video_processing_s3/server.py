from concurrent import futures
from datetime import datetime
import logging
import cv2
from minio import Minio
import grpc
import os
import helloworld_pb2
import helloworld_pb2_grpc
tmp = "/tmp/"

minioEnvKey = "MINIO_ADDRESS"
vid1_name = 'vid1.mp4'
vid2_name = 'vid2.mp4'
vid1_path = '/' + vid1_name
vid2_path = '/' + vid2_name

minioAddress = os.getenv(minioEnvKey)

def video_processing(video_path):
    result_file_path = tmp + video_path

    video = cv2.VideoCapture(video_path)

    width = int(video.get(3))
    height = int(video.get(4))

    fourcc = cv2.VideoWriter_fourcc(*'XVID')
    out = cv2.VideoWriter(result_file_path, fourcc, 20.0, (width, height))

    while video.isOpened():
        ret, frame = video.read()

        if ret:
            gray_frame = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
            tmp_file_path = tmp+'tmp.jpg'
            cv2.imwrite(tmp_file_path, gray_frame)
            gray_frame = cv2.imread(tmp_file_path)
            out.write(gray_frame)
        else:
            break

    video.release()
    out.release()
    return

responses = ["record_response", "replay_response"]

class Greeter(helloworld_pb2_grpc.GreeterServicer):

    def SayHello(self, request, context):
        logging.warning('NNS: SayHello execution started -- video_processing')
        start_time = datetime.now()
        # if minioAddress == None:
        #     return None

        # minioClient = Minio(minioAddress,
        #         access_key='minioadmin',
        #         secret_key='minioadmin',
        #         secure=False)

        if request.name == "record":
            msg = 'Hello, %s! -- video_processing -- ' % responses[0]

            # minioClient.fget_object('mybucket', vid1_name, vid1_path)
            video_processing(vid1_path)
        elif request.name == "replay":
            msg = 'Hello, %s! -- video_processing -- ' % responses[1]

            # minioClient.fget_object('mybucket', vid2_name, vid2_path)
            video_processing(vid2_path)
        else:
            msg = 'Hello, %s! -- video_processing -- ' % request.name

            # minioClient.fget_object('mybucket', vid1_name, vid1_path)
            video_processing(vid1_path)

        msg += str(datetime.now() - start_time)
        reply = helloworld_pb2.HelloReply(message=msg)
        logging.warning('NNS: SayHello execution ended -- video_processing')
        return reply


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    helloworld_pb2_grpc.add_GreeterServicer_to_server(Greeter(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    server.wait_for_termination()


if __name__ == '__main__':
    logging.basicConfig()
    serve()
