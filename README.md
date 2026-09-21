<p align="center">
  <img src=".github/banner.svg" width="100%" alt="5G-MAG Reference Tools, Content Delivery Protocols: CMMF Encoder">
</p>

<p align="center">
  Encodes packaged media into Coded Multisource Media Format bitstreams for multisource delivery,
  per ETSI TS 103 973.
</p>

<p align="center">
  <img alt="Status: experimental"
    src="https://img.shields.io/badge/Status-Experimental-c0392b">
  <a href="https://github.com/5G-MAG/rt-cmmf-encoder/releases"><img alt="Version"
    src="https://img.shields.io/github/v/release/5G-MAG/rt-cmmf-encoder?label=Version&sort=semver"></a>
  <a href="LICENSE"><img alt="License: 5G-MAG Public License v1.0"
    src="https://img.shields.io/badge/License-5G--MAG%20PL%20v1.0-blue"></a>
</p>

<p align="center">
  <a href="https://www.5g-mag.com/reference-tools/content-delivery">Project page</a> &nbsp;&middot;&nbsp;
  <a href="https://github.com/5G-MAG/rt-cmmf-encoder/issues">Issues</a> &nbsp;&middot;&nbsp;
  <a href="https://www.5g-mag.com/contributing">Contributing</a>
</p>

---

## At a glance

|  |  |
|---|---|
| **Implements** | ETSI TS 103 973 V1.2.2 (2026-05), *Coded Multisource Media Format (CMMF) for Content Distribution and Delivery* |
| **Code type** | xCD-1, the code type that specification defines normatively in its annex A |
| **Part of** | [Content Delivery Protocols](https://www.5g-mag.com/reference-tools/content-delivery), alongside [rt-libflute](https://github.com/5G-MAG/rt-libflute) |

## Introduction

Coded Multisource Media Format (CMMF) provides a generic container format that supports multimedia (e.g. video and audio streaming, broadcast, XR, video conferencing, and online gaming) delivery through coding the underlying content. This format supports multiple types of codes (currently xCD-1, RaptorQ, and Reed-Solomon) and can be optimized for a range of networks and use cases. Specifically, CMMF supports efficient decentralized multi-source and multi-path content delivery for use cases such as audio and video streaming that require high availability/robustness but also have strict latency and bandwidth constraints.

CMMF is designed to operate with existing and future streaming source formats (e.g. HLS, MPEG-DASH, CMAF, etc.) and network protocols (e.g. HTTP, TCP, UDP, WebRTC, etc.), while remaining protocol-agnostic. A multisource media encoder is envisioned to take an existing packaged media format as a source and generate CMMF bitstreams for delivery over networks to clients for rendering.

This repository is the encoder side: source data in, CMMF bitstreams out. The decoder side and the
delivery architecture are out of scope here.

## Specification

In this repository is a Go implementation of the [ETSI CMMF](https://www.etsi.org/deliver/etsi_ts/103900_103999/103973/01.02.02_60/ts_103973v010202p.pdf) 
encoder that can encode arbitrary data into a valid CMMF bitstream. 

There is also a command-line application that takes in some source text (as well as the path to an ETSI Encoder Config 
file) and writes out a file containing the encoded bitstream.

The latest version (as of this revision of the encoder) of the encoder configuration manifest/schema can be found 
[here](https://www.etsi.org/deliver/etsi_ts/103900_103999/103973/01.02.02_60/). 

## Install dependencies

Please download and install the Go binary release for your platform from [here](https://go.dev/dl/). 

Installation instructions can be found at https://go.dev/doc/install. 

## Usage

### Command-line example application

The `cmd` folder contains a command-line application that serves to illustrate the end-to-end workflow of
reading some arbitrary bytes from a source, and creating an output file containing the encoded bitstream.

Clone the repository and change into it:

```bash
git clone https://github.com/5G-MAG/rt-cmmf-encoder.git
cd rt-cmmf-encoder
```

Requires Go 1.26.2 or later (the version `go.mod` pins).

Build the application:

```bash
go build -o etsi_encoder ./cmd/etsi_encoder
```

List its options:

```bash
./etsi_encoder -h
```

```
Usage of ./etsi_encoder:
  -enc-config string
    	path to the encoder config file
  -enc-output string
    	path to the encoded output file
  -src-content string
    	path to the source content file
```

Encode the bundled example, a plain text file, using the bundled example configuration:

```bash
./etsi_encoder \
  -enc-config ./cmd/etsi_encoder/example_encoder_config.json \
  -src-content ./cmd/etsi_encoder/gutenberg_aliceinwonderland.txt \
  -enc-output ./out.dat
```

The application reads the first 16384 bytes of the source file and writes eight encoded packets, each 2053
bytes, behind a 56-byte bitstream header, giving a 16480-byte `out.dat`. Encoding is deterministic, so the
same source and configuration always produce the same bytes:

```bash
sha256sum ./out.dat
```

```
bfea6078c6c7ccb36ffccce464ee62970baaff9a1c47bb05afc3706cb0986f00  ./out.dat
```

To run without building a binary first, substitute `go run ./cmd/etsi_encoder` for `./etsi_encoder` in the
commands above:

```bash
go run ./cmd/etsi_encoder -h
```

### Unit tests

To execute the full set of unit tests in this repository:

```bash
go test ./...
```

Add `-v` to list each case as it runs:

```bash
go test -v ./...
```

### Documentation

Documentation for this package can be viewed by running:

```bash
go doc -http
```

This prints the address to open in a browser. The first run downloads the documentation server and can take a
few minutes before it prints anything.

## Implementation Notes 

As of this writing, only `code type 0`, `xCD-1` is supported. Please refer to `Annex A` in the technical specification 
for more details. 

In addition, not all of the optional features described in the technical spec have been implemented.  


## Contributing

Contributions are welcome. How to raise an issue, fork the repository and open a pull request, and
the Contributor License Agreement required before code can be merged, are described at
<https://www.5g-mag.com/contributing>.

## License

Distributed under the 5G-MAG Public License v1.0. See [LICENSE](LICENSE).
