# SDP
[RFC 4566](https://tools.ietf.org/html/rfc4566)
SDP: Session Description Protocol in go.

In active development.

## Parameters
```
Session description
   v=  (protocol version)
   o=  (originator and session identifier)
   s=  (session name)
   i=* (session information)
   u=* (URI of description)
   e=* (email address)
   p=* (phone number)
   c=* (connection information -- not required if included in
        all media)
   b=* (zero or more bandwidth information lines)
   One or more time descriptions ("t=" and "r=" lines; see below)
   z=* (time zone adjustments)
   k=* (encryption key)
   a=* (zero or more session attribute lines)
   Zero or more media descriptions

Time description
   t=  (time the session is active)
   r=* (zero or more repeat times)

Media description, if present
   m=  (media name and transport address)
   i=* (media title)
   c=* (connection information -- optional if included at
        session level)
   b=* (zero or more bandwidth information lines)
   k=* (encryption key)
   a=* (zero or more media attribute lines)
```

### Encoding
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

### Decoding
- [x] v (protocol version)
- [ ] o (originator and session identifier)
- [ ] s (session name)
- [ ] i (session information)
- [ ] u (URI of description)
- [ ] e (email address)
- [ ] p (phone number)
- [ ] c (connection information)
- [ ] b (zero or more bandwidth information lines)
- [ ] t (time)
- [ ] r (repeat)
- [ ] z (time zone adjustments)
- [ ] k (encryption key)
- [x] a (zero or more session attribute lines)

### TODO:
- [x] Encoding
- [x] Parsing
- [ ] Decoding
- [ ] High level encoding/decoding for Session and Media descriptions.

### Possible optimizations
There are comments `// ALLOCATIONS: suboptimal.` and `// CPU: suboptimal. `
that indicate suboptimal implementation that can be optimized. There are often
a benchmarks for this pieces.