# k8s API server watch tests

I got a few recent reports about AKS API server client watches being silently dropped, e.g. [this one](https://github.com/Azure/AKS/issues/1755). The indend of this repro is to develop a few tools and tests to repro the issue and make the debugging a bit easier.

## timer-sever
timer-server is a simple TCP server that sends time back to client on a timer. The server *intentionally* disabled TCP keepalive. The intend is to test the client behavior when Azure SLB sends TCP RST on idle timeout. 

The server has 2 ports, port 8005 sends data every 5 seconds, and port 8600 sends data every 600 seconds. The server deployment includes a tcpdump sidecar, so we can see the logs always.

### Test Results

| Client  | Result | 
|---|---|
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
|---|---|---|
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
