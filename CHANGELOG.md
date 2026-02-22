# Changelog

## [2.0.0](https://github.com/rpcpool/terraform-provider-lago/compare/v1.4.1...v2.0.0) (2026-02-22)


### ⚠ BREAKING CHANGES

* Removes internal Lago API client in favor of the official github.com/getlago/lago-go-client. This changes error handling and removes the internal client package. All resources now use the upstream Go client for API interactions.

### Features

* switch to github.com/getlago/lago-go-client ([9f0d5d5](https://github.com/rpcpool/terraform-provider-lago/commit/9f0d5d5fc4da487aa4cefece02988d09160fce61))

## [1.4.1](https://github.com/rpcpool/terraform-provider-lago/compare/v1.4.0...v1.4.1) (2026-02-20)


### Bug Fixes

* update module path and registry to rpcpool ([a2330be](https://github.com/rpcpool/terraform-provider-lago/commit/a2330be5599b195da5621b4a8b114d28cfb5d486))

## [1.4.0](https://github.com/rpcpool/terraform-provider-lago/compare/v1.3.0...v1.4.0) (2026-02-20)


### Features

* trigger release-please ([cd40468](https://github.com/rpcpool/terraform-provider-lago/commit/cd40468d9c2172c8ed34dcfccc606381a42d0e0c))

## [1.3.0](https://github.com/rpcpool/terraform-provider-lago/compare/v1.2.0...v1.3.0) (2026-02-20)


### Features

* trigger release-please ([067776d](https://github.com/rpcpool/terraform-provider-lago/commit/067776d91be7612314625e3f859b9b2f08fbb2c5))

## [1.2.0](https://github.com/rpcpool/terraform-provider-lago/compare/v1.1.0...v1.2.0) (2026-02-20)


### Features

* Merge branch 'main' of github.com:rpcpool/terraform-provider-lago ([875885f](https://github.com/rpcpool/terraform-provider-lago/commit/875885fb5dc58e799db91df53e93d76f4faf6b69))
* regenerate provider and resource documentation with tfplugindocs ([d31acd8](https://github.com/rpcpool/terraform-provider-lago/commit/d31acd8065ddf5562c1308f639c79ca9fd8fa9c2))

## [1.1.0](https://github.com/rpcpool/terraform-provider-lago/compare/v1.0.0...v1.1.0) (2026-02-20)


### Features

* trigger release-please ([408efa5](https://github.com/rpcpool/terraform-provider-lago/commit/408efa52379b51bfb760ce2289c1e29ed130f803))

## 1.0.0 (2026-02-19)


### Features

* Adopt best practice from hashicorp ([63152c9](https://github.com/rpcpool/terraform-provider-lago/commit/63152c9f0dede3f0175191c83a5b5a60817c1330))
* Initial commit ([7ff8917](https://github.com/rpcpool/terraform-provider-lago/commit/7ff8917ebb9d41c7add40f222cd62c9ec610314b))
* **provider:** add lago_id and filters_json fields to resources ([dd98ba2](https://github.com/rpcpool/terraform-provider-lago/commit/dd98ba2ff60175a4882a2859b02e384e57528fe5))
