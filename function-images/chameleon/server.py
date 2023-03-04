from concurrent import futures
from datetime import datetime
import logging

import grpc

import six
from chameleon import PageTemplate

import helloworld_pb2
import helloworld_pb2_grpc

BIGTABLE_ZPT = """\
<table xmlns="http://www.w3.org/1999/xhtml"
xmlns:tal="http://xml.zope.org/namespaces/tal">
<tr tal:repeat="row python: options['table']">
<td tal:repeat="c python: row.values()">
<span tal:define="d python: c + 1"
tal:attributes="class python: 'column-' + %s(d)"
tal:content="python: d" />
</td>
</tr>
</table>""" % six.text_type.__name__


responses = ["record_response", "replay_response"]

class Greeter(helloworld_pb2_grpc.GreeterServicer):

    def SayHello(self, request, context):
        logging.warning('NNS: SayHello execution started -- chameleon')
        start_time = datetime.now()
        tmpl = PageTemplate(BIGTABLE_ZPT)

        data = {}
        num_of_cols = 15
        num_of_rows = 10

        if request.name == "record":
            msg = 'Hello, %s! -- chameleon -- ' % responses[0]
            num_of_cols = 15
            num_of_rows = 10
        elif request.name == "replay":
            msg = 'Hello, %s! -- chameleon -- ' % responses[1]
            num_of_cols = 10
            num_of_rows = 15
        else:
            msg = 'Hello, %s! -- chameleon -- ' % request.name

        for i in range(num_of_cols):
            data[str(i)] = i

        table = [data for x in range(num_of_rows)]
        options = {'table': table}

        data = tmpl.render(options=options)
        msg += str(datetime.now() - start_time)
        reply = helloworld_pb2.HelloReply(message=msg)
        logging.warning('NNS: SayHello execution ended -- chameleon')
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
