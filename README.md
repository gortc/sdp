[![Build Status](https://travis-ci.org/ernado/sdp.svg?branch=master)](https://travis-ci.org/ernado/sdp)

# SDP
[RFC 4566](https://tools.ietf.org/html/rfc4566)
SDP: Session Description Protocol in go.

In alpha stage.

### Supported params
- [x] v (protocol version)
- [x] o (originator and session identifier)
- [x] s (session name)
- [x] i (session information)
- [x] u (URI of description)
- [x] e (email address)
- [x] p (phone number)
- [x] c (connection information)
- [x] b (zero or more bandwidth information lines)
- [x] t (time)
- [x] r (repeat)
- [x] z (time zone adjustments)
- [x] k (encryption key)
- [x] a (zero or more session attribute lines)
- [x] m (media name and transport address)

### TODO:
- [x] Encoding
- [x] Parsing
- [x] High level encoding
- [x] High level decoding
- [ ] Examples
- [x] CI
- [ ] Include to high-level CI

### Possible optimizations
There are comments `// ALLOCATIONS: suboptimal.` and `// CPU: suboptimal. `
that indicate suboptimal implementation that can be optimized. There are often
a benchmarks for this pieces.