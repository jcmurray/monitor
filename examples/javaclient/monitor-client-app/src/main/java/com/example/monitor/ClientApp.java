// cSpell.language:en-GB
// cSpell:disable

package com.example.monitor;

import io.grpc.ManagedChannelBuilder;
import io.grpc.ManagedChannel;
import com.example.monitor.clientapi.*;
import com.google.protobuf.Empty;

import java.util.concurrent.TimeUnit;
import java.util.concurrent.CountDownLatch;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;
import java.util.Iterator;
import java.util.List;

public class ClientApp {

    private static final Logger LOG = Logger.getLogger(ClientApp.class.getName());
    private final ManagedChannel channel;
    private final ClientServiceGrpc.ClientServiceBlockingStub blockingStub;
    private final ClientServiceGrpc.ClientServiceStub asyncStub;

    public ClientApp(String host, int port) {
        this(ManagedChannelBuilder.forAddress(host, port).usePlaintext().build());
    }

    ClientApp(ManagedChannel channel) {
        this.channel = channel;
        blockingStub = ClientServiceGrpc.newBlockingStub(channel);
        asyncStub = ClientServiceGrpc.newStub(channel);
    }

    public void shutdown() throws InterruptedException {
        channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
    }

    public void sendTextMessage(String message, String forUser) {
        LOG.info("Will try to send message: " + message + ", to: " + forUser);

        TextMessage textMessage = TextMessage.newBuilder().setFor(forUser).setMessage(message).build();
        TextMessageResponse response;

        try {
            response = blockingStub.sendTextMessage(textMessage);
        } catch (StatusRuntimeException e) {
            LOG.warn("RPC failed: " + e.getStatus());
            return;
        }
        LOG.info("Response: " + response.getSuccess() + ", message: " + response.getMessage());
    }

    public void statusSync() {
        LOG.info("Querying status of server using Sync API");
        Iterator<WorkerDetails> wds;
        try {
            wds = blockingStub.status(Empty.getDefaultInstance());
            for (int i = 1; wds.hasNext(); i++) {
                WorkerDetails wd = wds.next();
                LOG.info("Name: " + wd.getName() + ", Id: " + wd.getId());
            }
        } catch (StatusRuntimeException e) {
            LOG.warn("RPC failed: " + e.getStatus());
        }
    }

    public CountDownLatch statusAsync() {
        LOG.info("Querying status of server using Async API");
        final CountDownLatch finishLatch = new CountDownLatch(1);

        try {
            asyncStub.status(Empty.getDefaultInstance(), new StreamObserver<WorkerDetails>() {
                @Override
                public void onNext(WorkerDetails wds) {
                    LOG.info("Name: " + wds.getName() + ", Id: " + wds.getId());
                    List<Subscription> subs = wds.getWorkerSubscriptionList();
                    for (Subscription sub : subs) {
                        LOG.info("\tId: " + sub.getId() + ", Label: " + sub.getLabel());
                    }
                }

                @Override
                public void onError(Throwable t) {
                    LOG.warn("RPC error: " + t.getMessage());
                    finishLatch.countDown();
                }

                @Override
                public void onCompleted() {
                    LOG.info("Completed");
                    finishLatch.countDown();
                }
            });
        } catch (StatusRuntimeException e) {
            LOG.warn("RPC failed: " + e.getStatus());
        }
        return finishLatch;
    }

    public static void main(String[] args) throws Exception {
        ClientApp client = new ClientApp("localhost", 9998);
        try {
            client.sendTextMessage("Hello World!", "");
            client.statusSync();
            CountDownLatch finishLatch = client.statusAsync();
            if (!finishLatch.await(1, TimeUnit.MINUTES)) {
                LOG.warn("statusSync not finished within 1 minute");
            }
        } finally {
            client.shutdown();
        }
    }
}
