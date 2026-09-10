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
  <a href="https://www.5g-mag.com/community/contributing">Contributing</a>
</p>

---

## At a glance

|  |  |
|---|---|
| **Implements** | ETSI TS 103 973 V1.1.1 (2024-10), *Coded Multisource Media Format (CMMF) for Content Distribution and Delivery* |
| **Code type** | xCD-1, the code type that specification defines normatively in its annex A |
| **Part of** | [Content Delivery Protocols](https://www.5g-mag.com/reference-tools/content-delivery), alongside [rt-libflute](https://github.com/5G-MAG/rt-libflute) |

## Introduction

CMMF is a container format for carrying coded media from more than one source. ETSI TS 103 973
V1.1.1, clause 1: "The present document specifies a Coded Multisource Media Format (CMMF) container."

It supplements existing packaging rather than replacing it: an encoder takes an already packaged
media format, such as ISO BMFF or CMAF, and produces coded bitstreams that can be delivered from
several sources at once. Because CMMF carries no manifest of its own, bitstreams for an asset can be
created or discarded without touching the others.

This repository is the encoder side: source data in, CMMF bitstreams out. The decoder side and the
delivery architecture are out of scope here.

## Specification

Built against **ETSI TS 103 973 V1.1.1 (2024-10)**, a version rather than a release name.

Clause-by-clause coverage, and what is still absent, is recorded on the project page rather than
here: <https://www.5g-mag.com/reference-tools/content-delivery>

This is the repository's initial commit, so no encoder code has landed yet. Install, build and usage
instructions follow with the first code.

## Getting the source

```bash
cd ~
git clone https://github.com/5G-MAG/rt-cmmf-encoder.git
```

## Contributing

Contributions are welcome. How to raise an issue, fork the repository and open a pull request, and
the Contributor License Agreement required before code can be merged, are described at
<https://www.5g-mag.com/community/contributing>.

## License

Distributed under the 5G-MAG Public License v1.0. See [LICENSE](LICENSE).
