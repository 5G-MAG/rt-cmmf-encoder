<p align="center">
  <img src=".github/banner.svg" width="100%" alt="5G-MAG Reference Tools, Multimedia Delivery Protocols: CMMF Encoder">
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
  <a href="https://www.5g-mag.com/reference-tools/multimedia">Project page</a> &nbsp;&middot;&nbsp;
  <a href="https://github.com/5G-MAG/rt-cmmf-encoder/issues">Issues</a> &nbsp;&middot;&nbsp;
  <a href="https://www.5g-mag.com/community/contributing">Contributing</a>
</p>

---

## At a glance

|  |  |
|---|---|
| **Implements** | ETSI TS 103 973 V1.1.1 (2024-10), *Coded Multisource Media Format (CMMF) for Content Distribution and Delivery* |
| **Code type** | xCD-1, the code type that specification defines normatively in its annex A |
| **Part of** | [Multimedia Delivery Protocols](https://www.5g-mag.com/reference-tools/multimedia), alongside [rt-libflute](https://github.com/5G-MAG/rt-libflute) |

## Introduction

CMMF is a container format for carrying coded media from more than one source. ETSI TS 103 973
V1.1.1, clause 1: "The present document specifies a Coded Multisource Media Format (CMMF) container."

It supplements existing packaging rather than replacing it: an encoder takes an already packaged
media format, such as ISO BMFF or CMAF, and produces coded bitstreams that can be delivered from
several sources at once. Because CMMF carries no manifest of its own, bitstreams for an asset can be
created or discarded without touching the others.

This repository is the encoder side: source data in, CMMF bitstreams out. The decoder side and the
delivery architecture are out of scope here.

## Specification and conformance

Built against **ETSI TS 103 973 V1.1.1 (2024-10)**. A version, not a release name: a reader checking
against a different version of the same document will disagree for no reason.

Which clauses are implemented, partially implemented or still absent is recorded on the project
page, not here: <https://www.5g-mag.com/reference-tools/multimedia>. The repository carries the
code; the conformance record is the reader-facing statement of what that code does and does not do,
and it belongs where readers look for it.

## Implementation status

This is the repository's initial commit. No encoder code has landed yet, which is what the
Experimental status badge above reports. Install, build and usage instructions will be added with
the first code, and the conformance record on the project page will be updated as clauses land.

## Getting the source

```bash
cd ~
git clone https://github.com/5G-MAG/rt-cmmf-encoder.git
```

## Contributing

Raise an issue, fork the repository, work on a branch, open a pull request. The full flow, and the
Contributor License Agreement required before code can be merged, are described at
<https://www.5g-mag.com/community/contributing>. Releases follow
<https://www.5g-mag.com/community/release-process>.

## License

Distributed under the 5G-MAG Public License v1.0. See [LICENSE](LICENSE), and
<https://www.5g-mag.com/license> for what that licence permits and the third-party dependencies it
does not cover.
