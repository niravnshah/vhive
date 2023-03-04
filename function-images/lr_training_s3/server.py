from concurrent import futures
from datetime import datetime
import logging
import os
import grpc
from minio import Minio
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
import joblib
import pandas as pd
import re
import io

import helloworld_pb2
import helloworld_pb2_grpc

cleanup_re = re.compile('[^a-z]+')

responses = ["record_response", "replay_response"]

def cleanup(sentence):
    sentence = sentence.lower()
    sentence = cleanup_re.sub(' ', sentence).strip()
    return sentence

minioEnvKey = "MINIO_ADDRESS"
df_name = 'dataset.csv'
df2_name = 'dataset2.csv'
df_path = '/' + df_name
df2_path = '' + df2_name

minioAddress = os.getenv(minioEnvKey)

class Greeter(helloworld_pb2_grpc.GreeterServicer):

    def SayHello(self, request, context):
        logging.basicConfig()
        logging.warning('NNS: SayHello execution started -- lr_training')
        start_time = datetime.now()
        # if minioAddress == None:
        #     return None

        # minioClient = Minio(minioAddress,
        #         access_key='minioadmin',
        #         secret_key='minioadmin',
        #         secure=False)

        if request.name == "record":
            msg = 'Hello, %s! -- lr_training -- ' % responses[0]
            # minioClient.fget_object('mybucket', df_name, df_path)

            df = pd.read_csv(df_path)
            df['train'] = df['Text'].apply(cleanup)
            tfidf_vector = TfidfVectorizer(min_df=100).fit(df['train'])
            train = tfidf_vector.transform(df['train'])
            model = LogisticRegression()
            model.fit(train, df['Score'])
        elif request.name == "replay":
            msg = 'Hello, %s! -- lr_training -- ' % responses[1]
            # minioClient.fget_object('mybucket', df2_name, df2_path)

            df2 = pd.read_csv(df2_path)
            df2['train'] = df2['Text'].apply(cleanup)
            tfidf_vector2 = TfidfVectorizer(min_df=100).fit(df2['train'])
            train2 = tfidf_vector2.transform(df2['train'])
            model2 = LogisticRegression()
            model2.fit(train2, df2['Score'])
        else:
            msg = 'Hello, %s! -- lr_training -- ' % request.name
            # minioClient.fget_object('mybucket', df_name, df_path)

            df = pd.read_csv(df_path)
            df['train'] = df['Text'].apply(cleanup)
            tfidf_vector = TfidfVectorizer(min_df=100).fit(df['train'])
            train = tfidf_vector.transform(df['train'])
            model = LogisticRegression()
            model.fit(train, df['Score'])

        #joblib.dump(model, '/var/local/dir/lr_model.pk')
        msg += str(datetime.now() - start_time)
        reply = helloworld_pb2.HelloReply(message=msg)
        logging.warning('NNS: SayHello execution ended -- lr_training')
        return reply


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    helloworld_pb2_grpc.add_GreeterServicer_to_server(Greeter(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    server.wait_for_termination()


if __name__ == '__main__':
    serve()
