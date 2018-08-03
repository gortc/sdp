# E2E

End-to-end tests.

1. Open browser
2. Navigate to page
3. Get SDP
4. Do POST to provided URL

SDP example:
```
v=0
o=- 6224905574295094272 2 IN IP4 127.0.0.1
s=-
t=0 0
a=group:BUNDLE data
a=msid-semantic: WMS
m=application 9 DTLS/SCTP 5000
c=IN IP4 0.0.0.0
a=ice-ufrag:9s9d
a=ice-pwd:tFb7JsbaCI4U0cWAIv/EZiuH
a=ice-options:trickle
a=fingerprint:sha-256 AA:A2:06:BB:65:B3:67:E6:A9:2E:3A:70:73:DD:D4:6D:25:8E:49:44:18:CE:F1:4C:8D:A7:8B:0D:97:6B:24:62
a=setup:actpass
a=mid:data
a=sctpmap:5000 webrtc-datachannel 1024

```