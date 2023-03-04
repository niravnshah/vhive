from concurrent import futures
from datetime import datetime
import logging
import os
import grpc

from PIL import Image, ImageOps

import helloworld_pb2
import helloworld_pb2_grpc

from minio import Minio

minioEnvKey = "MINIO_ADDRESS"
image_name = 'img2.jpeg'
image2_name = 'img3.jpeg'
image_path = '/' + image_name
image_path2 = '/' +image2_name

responses = ["record_response", "replay_response"]

minioAddress = os.getenv(minioEnvKey)

class Greeter(helloworld_pb2_grpc.GreeterServicer):

    def SayHello(self, request, context):
        logging.warning('NNS: SayHello execution started -- image_rotate')
        start_time = datetime.now()
        # if minioAddress == None:
        #     return None

        # minioClient = Minio(minioAddress,
        #         access_key='minioadmin',
        #         secret_key='minioadmin',
        #         secure=False)
        if request.name == "record":
            msg = 'Hello, %s! -- image_rotate -- ' % responses[0]
            # minioClient.fget_object('mybucket', image_name, image_path)
            image = Image.open(image_path)
            img = image.transpose(Image.ROTATE_90)
        elif request.name == "replay":
            msg = 'Hello, %s! -- image_rotate -- ' % responses[1]
            # minioClient.fget_object('mybucket', image2_name, image_path2)
            image2 = Image.open(image_path2)
            img = image2.transpose(Image.ROTATE_90)
        else:
            msg = 'Hello, %s! -- image_rotate -- ' % request.name
            # minioClient.fget_object('mybucket', image_name, image_path)
            image = Image.open(image_path)
            img = image.transpose(Image.ROTATE_90)

        msg += str(datetime.now() - start_time)
        reply = helloworld_pb2.HelloReply(message=msg)
        logging.warning('NNS: SayHello execution ended -- image_rotate')
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
