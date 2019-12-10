// cSpell.language:en-GB
// cSpell:disable

var PROTO_PATH = __dirname + '../../../protocols/clientprotocol.proto';

var grpc = require('grpc');
var protoLoader = require('@grpc/proto-loader');
var packageDefinition = protoLoader.loadSync(
	PROTO_PATH, {
		keepCase: true,
		longs: String,
		enums: String,
		defaults: true,
		oneofs: true
	}
);
var clientapi_proto = grpc.loadPackageDefinition(packageDefinition).clientapi;

function main() {
	var client = new clientapi_proto.ClientService('localhost:9998',
		grpc.credentials.createInsecure());
	var callStatus = client.Status({});

	client.SendTextMessage({
		message: "Hello World!",
		for: ""
	}, function (err, response) {
		console.log('Response: ', response);
	});

	callStatus.on('data', function (wds) {
		console.log('data called: ', wds);
	});
	callStatus.on('end', function () {
		console.log('end received')
	});
	callStatus.on('error', function (e) {
		console.log('error received: ', e)
	});
	callStatus.on('status', function (status) {
		console.log('status received: ', status)
	});

}

main();