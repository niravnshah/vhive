# Copyright 2015 gRPC authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""The Python implementation of the GRPC helloworld.Greeter server."""

from concurrent import futures
from datetime import datetime
import logging
import os
import pickle
import numpy as np
import torch
import rnn
import string

import grpc

import helloworld_pb2
import helloworld_pb2_grpc

torch.set_num_threads(1)

responses = ["record_response", "replay_response"]

language = 'Scottish'
language2 = 'Russian'
start_letters = 'ABCDEFGHIJKLMNOP'
start_letters2 = 'QRSTUVWXYZABCDEF'

with open('/rnn_params.pkl', 'rb') as pkl:
    params = pickle.load(pkl)

all_categories =['French', 'Czech', 'Dutch', 'Polish', 'Scottish', 'Chinese', 'English', 'Italian', 'Portuguese', 'Japanese', 'German', 'Russian', 'Korean', 'Arabic', 'Greek', 'Vietnamese', 'Spanish', 'Irish']
n_categories = len(all_categories)
all_letters = string.ascii_letters + " .,;'-"
n_letters = len(all_letters) + 1

rnn_model = rnn.RNN(n_letters, 128, n_letters, all_categories, n_categories, all_letters, n_letters)
rnn_model.load_state_dict(torch.load('/rnn_model.pth'))
rnn_model.eval()

class Greeter(helloworld_pb2_grpc.GreeterServicer):

    def SayHello(self, request, context):
        logging.basicConfig()
        logging.warning('NNS: SayHello execution started -- rnn_serving')
        start_time = datetime.now()
        if request.name == "record":
            msg = 'Hello, %s! -- rnn_serving -- ' % responses[0]
            output_names = list(rnn_model.samples(language, start_letters))
        elif request.name == "replay":
            msg = 'Hello, %s! -- rnn_serving -- ' % responses[1]
            output_names = list(rnn_model.samples(language2, start_letters2))
        else:
            msg = 'Hello, %s! -- rnn_serving -- ' % request.name
            output_names = list(rnn_model.samples(language, start_letters))

        msg += str(datetime.now() - start_time)
        reply = helloworld_pb2.HelloReply(message=msg)
        logging.warning('NNS: SayHello execution ended -- rnn_serving')
        return reply


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    helloworld_pb2_grpc.add_GreeterServicer_to_server(Greeter(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    server.wait_for_termination()


if __name__ == '__main__':
    serve()
