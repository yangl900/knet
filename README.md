# k8s API server watch tests

I got a few recent reports about AKS API server client watches being silently dropped, e.g. [this one](https://github.com/Azure/AKS/issues/1755). The indend of this repro is to develop a few tools and tests to repro the issue and make the debugging a bit easier.

## timer-sever
timer-server is a simple TCP server that sends time back to client on a timer. The server *intentionally* disabled TCP keepalive. The intend is to test the client behavior when Azure SLB sends TCP RST on idle timeout. 

The server has 2 ports, port 8005 sends data every 5 seconds, and port 8600 sends data every 600 seconds. The server deployment includes a tcpdump sidecar, so we can see the logs always.

### Test Results

| Client  | Result | 
|---|---|---|
|  curl | handles RST well  |  
|  python | TODO |
|  .Net Core | TODO | 

### `curl`
> Test environment: AKS cluster v1.17.11 with Standard LB

Curl handles the RST and returns error `curl: (56) Recv failure: Connection reset by peer`

client side command and output:
```bash
> curl 20.51.xx.x:8600 --no-keepalive --http0.9 || date -u
Hello! Current time is 2020-10-26T06:02:01Z. I'm going to send message every 600 seconds. No KeepAlive in this TCP connection.
curl: (56) Recv failure: Connection reset by peer
Mon 26 Oct 2020 06:06:24 AM UTC
```

Server side tcpdump shows SLB sends RST on 4 min and 20s idle to server side too.

```bash
05:47:48.986286 IP 67.185.98.118.51648 > 10.244.0.49.8600: Flags [S], seq 3269278141, win 64240, options [mss 1460,sackOK,TS val 2594862467 ecr 0,nop,wscale 7], length 0
05:47:48.986305 IP 10.244.0.49.8600 > 67.185.98.118.51648: Flags [S.], seq 1850028384, ack 3269278142, win 65160, options [mss 1460,sackOK,TS val 2817406417 ecr 2594862467,nop,wscale 7], length 0
05:47:49.016881 IP 67.185.98.118.51648 > 10.244.0.49.8600: Flags [.], ack 1, win 502, options [nop,nop,TS val 2594862488 ecr 2817406417], length 0
05:47:49.016906 IP 67.185.98.118.51648 > 10.244.0.49.8600: Flags [P.], seq 1:80, ack 1, win 502, options [nop,nop,TS val 2594862489 ecr 2817406417], length 79
05:47:49.016912 IP 10.244.0.49.8600 > 67.185.98.118.51648: Flags [.], ack 80, win 509, options [nop,nop,TS val 2817406448 ecr 2594862489], length 0
05:47:49.017077 IP 10.244.0.49.8600 > 67.185.98.118.51648: Flags [P.], seq 1:128, ack 80, win 509, options [nop,nop,TS val 2817406448 ecr 2594862489], length 127
05:47:49.037529 IP 67.185.98.118.51648 > 10.244.0.49.8600: Flags [.], ack 128, win 502, options [nop,nop,TS val 2594862518 ecr 2817406448], length 0
05:52:11.667639 IP 67.185.98.118.51648 > 10.244.0.49.8600: Flags [R], seq 3269278221, win 0, length 0
```

### Python
TODO

### .Net Core
TODO

## apiserver-watcher
apiserver-watcher is a test pod implmemented using standard k8s client. It has 2 very simple go routines: `Watcher` and `Updater`. Watcher establish a watch connection to API server for all configmap changes in namespace. Updater will update a configmap named `trigger` every 600s (interval configurable). Whenver watcher got a `trigger` configmap update event, it will notify Updater. And Updater will wait for up to 10s for the Watcher notify, if it did not get the notify in 10s, it declares a test failure and log a new configmap named `failure-xxxx` with timestamp in it.

### Test Results
> Test server: AKS cluster v1.17.11 with Standard LB <br>

Test matrix

| Client  | Public cluster | Private cluster |
|---|---|---|---|
|  client-go | PASS |  TODO |
|  client-python | TODO | TODO |
|  .Net Core | TODO | TODO |

#### client-go tests
> [client-go](https://github.com/kubernetes/client-go) v0.19.3

The k8s client-go by default turns on TCP Keepalive, and the client side will send an ACK packet to API server every 30s. With this, even though the SLB default timeout is 4 minutes, the TCP connection will never be idle and so will never be reset.

The test ran for 24 hours and did not generate a single failure.

client side tcp dump showed the keep alive ACK every 30s.
```bash
06:26:09.341954 IP 10.244.0.62.43848 > 10.0.0.1.443: Flags [.], ack 181169, win 501, options [nop,nop,TS val 891891695 ecr 767313160], length 0
06:26:09.343035 IP 10.0.0.1.443 > 10.244.0.62.43848: Flags [.], ack 58300, win 501, options [nop,nop,TS val 767343881 ecr 891368331], length 0
06:26:14.461969 ARP, Request who-has 10.244.0.62 tell 10.244.0.1, length 28
06:26:14.461979 ARP, Reply 10.244.0.62 is-at 22:4d:49:5c:06:c6, length 28
06:26:40.061965 IP 10.244.0.62.43848 > 10.0.0.1.443: Flags [.], ack 181169, win 501, options [nop,nop,TS val 891922415 ecr 767343881], length 0
06:26:40.062616 IP 10.0.0.1.443 > 10.244.0.62.43848: Flags [.], ack 58300, win 501, options [nop,nop,TS val 767374600 ecr 891368331], length 0
06:26:45.181951 ARP, Request who-has 10.244.0.1 tell 10.244.0.62, length 28
06:26:45.182022 ARP, Reply 10.244.0.1 is-at 3a:3e:fa:e8:7b:0f, length 28
06:27:10.781939 IP 10.244.0.62.43848 > 10.0.0.1.443: Flags [.], ack 181169, win 501, options [nop,nop,TS val 891953135 ecr 767374600], length 0
06:27:10.782393 IP 10.0.0.1.443 > 10.244.0.62.43848: Flags [.], ack 58300, win 501, options [nop,nop,TS val 767405320 ecr 891368331], length 0

```

### client-python tests

#### Reuse connection with polling

> The test scenairo is simply make a request to API server, then sleep 10min, and make a request to API server again. Expectation is the TCP connection initalited by the first request should have been reset and not used by the second one.

#### Findings

1. Client python does NOT have TCP keepalive true by default. If server side does not send RST, stale connection will be left open. Related issue: 
   * https://github.com/kubernetes-client/python/issues/1234
   * https://github.com/kubernetes-client/python/issues/1158
   * https://github.com/kubernetes-client/python/issues/928
1. Client python does handle RST and close the connection.

Since there is no ACK on RST packet, it is still a concern without keepalive. If RST packet was not delivered, the connection could still be stale.

Following TCP dump showed that client-python handled RST and closed the connection, however, there is no keepalive. The connection was reset on `05:23:54.212827`. The next request initiated a new TCP connection on `05:29:31.672603`.
```bash
05:19:31.551172 IP 10.244.0.74.54246 > 10.0.0.1.443: Flags [S], seq 105636032, win 64240, options [mss 1460,sackOK,TS val 168766782 ecr 0,nop,wscale 7], length 0
05:19:31.552363 IP 10.0.0.1.443 > 10.244.0.74.54246: Flags [S.], seq 878887080, ack 105636033, win 65160, options [mss 1440,sackOK,TS val 849746091 ecr 168766782,nop,wscale 7], length 0
05:19:31.552381 IP 10.244.0.74.54246 > 10.0.0.1.443: Flags [.], ack 1, win 502, options [nop,nop,TS val 168766783 ecr 849746091], length 0
05:19:31.554711 IP 10.244.0.74.54246 > 10.0.0.1.443: Flags [P.], seq 1:518, ack 1, win 502, options [nop,nop,TS val 168766785 ecr 849746091], length 517
05:19:31.555483 IP 10.0.0.1.443 > 10.244.0.74.54246: Flags [.], ack 518, win 506, options [nop,nop,TS val 849746094 ecr 168766785], length 0
05:19:31.570099 IP 10.0.0.1.443 > 10.244.0.74.54246: Flags [P.], seq 1:2424, ack 518, win 506, options [nop,nop,TS val 849746108 ecr 168766785], length 2423
05:19:31.570105 IP 10.244.0.74.54246 > 10.0.0.1.443: Flags [.], ack 2424, win 498, options [nop,nop,TS val 168766801 ecr 849746108], length 0
05:19:31.570846 IP 10.244.0.74.54246 > 10.0.0.1.443: Flags [P.], seq 518:628, ack 2424, win 501, options [nop,nop,TS val 168766801 ecr 849746108], length 110
05:19:31.571110 IP 10.244.0.74.54246 > 10.0.0.1.443: Flags [P.], seq 628:2115, ack 2424, win 501, options [nop,nop,TS val 168766802 ecr 849746108], length 1487
05:19:31.571273 IP 10.0.0.1.443 > 10.244.0.74.54246: Flags [.], ack 628, win 506, options [nop,nop,TS val 849746110 ecr 168766801], length 0
05:19:31.571421 IP 10.0.0.1.443 > 10.244.0.74.54246: Flags [.], ack 2115, win 501, options [nop,nop,TS val 849746110 ecr 168766802], length 0
05:19:31.571502 IP 10.0.0.1.443 > 10.244.0.74.54246: Flags [P.], seq 2424:2592, ack 2115, win 501, options [nop,nop,TS val 849746110 ecr 168766802], length 168
05:19:31.575764 IP 10.0.0.1.443 > 10.244.0.74.54246: Flags [P.], seq 2592:3213, ack 2115, win 501, options [nop,nop,TS val 849746114 ecr 168766802], length 621
05:19:31.575793 IP 10.244.0.74.54246 > 10.0.0.1.443: Flags [.], ack 3213, win 501, options [nop,nop,TS val 168766806 ecr 849746110], length 0
05:23:54.212827 IP 10.0.0.1.443 > 10.244.0.74.54246: Flags [R], seq 878890293, win 0, length 0
05:29:31.672603 IP 10.244.0.74.59098 > 10.0.0.1.443: Flags [S], seq 2677600961, win 64240, options [mss 1460,sackOK,TS val 169366903 ecr 0,nop,wscale 7], length 0
05:29:31.673293 IP 10.0.0.1.443 > 10.244.0.74.59098: Flags [S.], seq 1890366456, ack 2677600962, win 65160, options [mss 1440,sackOK,TS val 850346211 ecr 169366903,nop,wscale 7], length 0
05:29:31.673309 IP 10.244.0.74.59098 > 10.0.0.1.443: Flags [.], ack 1, win 502, options [nop,nop,TS val 169366904 ecr 850346211], length 0
05:29:31.673811 IP 10.244.0.74.59098 > 10.0.0.1.443: Flags [P.], seq 1:518, ack 1, win 502, options [nop,nop,TS val 169366904 ecr 850346211], length 517
05:29:31.674435 IP 10.0.0.1.443 > 10.244.0.74.59098: Flags [.], ack 518, win 506, options [nop,nop,TS val 850346212 ecr 169366904], length 0
```

