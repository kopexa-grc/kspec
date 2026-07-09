# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.4](https://github.com/kopexa-grc/kspec/compare/v0.2.3...v0.2.4) (2026-03-30)


### Features

* migrate CI back to self-hosted runners (public repo access enabled) ([6e439e2](https://github.com/kopexa-grc/kspec/commit/6e439e29c5a3fb2f4fea9324ab18499b1ccbbb36))


### Bug Fixes

* remove duplicate timeout-minutes ([d7a2330](https://github.com/kopexa-grc/kspec/commit/d7a23300dbee66b343245b66982ecc8c759d5a3b))
* remove setup-go (pre-installed on runner), lint gates all jobs ([edb7845](https://github.com/kopexa-grc/kspec/commit/edb7845341e83821577397c52e5808fefee61442))
* revert to GitHub-hosted runners (public repo runner group issue) ([862a1f2](https://github.com/kopexa-grc/kspec/commit/862a1f2c2b3be769d0bfb2ba1336ee04fa516076))
* revert to GitHub-hosted runners, remove runner-test ([55fa51b](https://github.com/kopexa-grc/kspec/commit/55fa51b9e65b9b5f90e68cbfe4bf3e5960aa1114))
* run lint on large runner (timeout on 2 vCPU) ([7553900](https://github.com/kopexa-grc/kspec/commit/7553900b495c51bac2637754937637a195a62ad3))
* run race tests on ubuntu-4c (needs CGO/gcc) ([c2f9535](https://github.com/kopexa-grc/kspec/commit/c2f9535310782ac0e8f5413cbf840707b59969ae))


### Dependencies

* **deps:** bump github.com/aws/aws-sdk-go-v2 from 1.41.4 to 1.41.5 ([#171](https://github.com/kopexa-grc/kspec/issues/171)) ([b54b009](https://github.com/kopexa-grc/kspec/commit/b54b0094429dc6eb93fa1b6ed08021f00a678c79))
* **deps:** bump github.com/aws/aws-sdk-go-v2/config ([#110](https://github.com/kopexa-grc/kspec/issues/110)) ([65942cf](https://github.com/kopexa-grc/kspec/commit/65942cff44724e9d964886e1090ab802a3a5a5c3))
* **deps:** bump github.com/aws/aws-sdk-go-v2/config ([#140](https://github.com/kopexa-grc/kspec/issues/140)) ([82c54d0](https://github.com/kopexa-grc/kspec/commit/82c54d0d599c97ac77da31a64556461e45f4e5a7))
* **deps:** bump github.com/aws/aws-sdk-go-v2/config ([#181](https://github.com/kopexa-grc/kspec/issues/181)) ([3ca3ef4](https://github.com/kopexa-grc/kspec/commit/3ca3ef4c6869b925dbee579574586e323eea301d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/config ([#97](https://github.com/kopexa-grc/kspec/issues/97)) ([08e5d9a](https://github.com/kopexa-grc/kspec/commit/08e5d9a69f9d5d41da647db622ca6ca3f9dbe64f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/credentials ([#117](https://github.com/kopexa-grc/kspec/issues/117)) ([5f33eda](https://github.com/kopexa-grc/kspec/commit/5f33edad28eb6d395a454a5f21f928b5a1513588))
* **deps:** bump github.com/aws/aws-sdk-go-v2/credentials ([#175](https://github.com/kopexa-grc/kspec/issues/175)) ([ee96b10](https://github.com/kopexa-grc/kspec/commit/ee96b10d3bfac02c4e3d34f61b460465cdf023de))
* **deps:** bump github.com/aws/aws-sdk-go-v2/credentials ([#93](https://github.com/kopexa-grc/kspec/issues/93)) ([fcbf0b3](https://github.com/kopexa-grc/kspec/commit/fcbf0b316cf837edb4c0aa59bce4b490dc3b5b1f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/acm ([#153](https://github.com/kopexa-grc/kspec/issues/153)) ([33bce22](https://github.com/kopexa-grc/kspec/commit/33bce22bc6048237a5735a7ff73e4b6d4325bae3))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/acm ([#182](https://github.com/kopexa-grc/kspec/issues/182)) ([41085ba](https://github.com/kopexa-grc/kspec/commit/41085ba5327b7e4910218c0d726f7ee317ce9b40))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/apigateway ([#108](https://github.com/kopexa-grc/kspec/issues/108)) ([79457cd](https://github.com/kopexa-grc/kspec/commit/79457cd5df872ac6a4719fd417540f6f23d7eab3))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/apigateway ([#146](https://github.com/kopexa-grc/kspec/issues/146)) ([d92e2b4](https://github.com/kopexa-grc/kspec/commit/d92e2b47cd41453fb5222ad9ce624c0031d1df99))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/apigateway ([#193](https://github.com/kopexa-grc/kspec/issues/193)) ([600f731](https://github.com/kopexa-grc/kspec/commit/600f7311291df606cdd717628bde445ea94c3f9a))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/apigatewayv2 ([#112](https://github.com/kopexa-grc/kspec/issues/112)) ([e39e8ee](https://github.com/kopexa-grc/kspec/commit/e39e8ee296529061cf7bd8629186d29ef07428fa))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/apigatewayv2 ([#155](https://github.com/kopexa-grc/kspec/issues/155)) ([f9f3cf5](https://github.com/kopexa-grc/kspec/commit/f9f3cf5eec26c1223ac6fbaf20e03c27bbe6cf1a))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/apigatewayv2 ([#166](https://github.com/kopexa-grc/kspec/issues/166)) ([e7f98f9](https://github.com/kopexa-grc/kspec/commit/e7f98f971bac5a83c97bccf1323ce56b20be94d2))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/autoscaling ([#145](https://github.com/kopexa-grc/kspec/issues/145)) ([274fd6b](https://github.com/kopexa-grc/kspec/commit/274fd6b69d58768aeb6c6be62a7ff179830d13d8))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/autoscaling ([#190](https://github.com/kopexa-grc/kspec/issues/190)) ([391f6df](https://github.com/kopexa-grc/kspec/commit/391f6df9bb015e29f3d6460dde4b3188365a6f3b))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudfront ([#128](https://github.com/kopexa-grc/kspec/issues/128)) ([be3172d](https://github.com/kopexa-grc/kspec/commit/be3172d3e82c7145695012735a79b4ef0b477e8c))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudfront ([#194](https://github.com/kopexa-grc/kspec/issues/194)) ([2bafcb9](https://github.com/kopexa-grc/kspec/commit/2bafcb9af16146588b72904a5aa4ae0558316c3e))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudtrail ([#120](https://github.com/kopexa-grc/kspec/issues/120)) ([c5522e1](https://github.com/kopexa-grc/kspec/commit/c5522e1f5df0867e8fd51629a1de1599340884f0))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudtrail ([#136](https://github.com/kopexa-grc/kspec/issues/136)) ([33bbb24](https://github.com/kopexa-grc/kspec/commit/33bbb24ad0a82fdc166f358f80b5905449bbe3dd))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudtrail ([#165](https://github.com/kopexa-grc/kspec/issues/165)) ([32bfbc6](https://github.com/kopexa-grc/kspec/commit/32bfbc61a96d6eddf5a4034dbab9688e00afdd08))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatch ([#119](https://github.com/kopexa-grc/kspec/issues/119)) ([2baf108](https://github.com/kopexa-grc/kspec/commit/2baf10895691689dc070981f3539ff9d9d8e45c0))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatch ([#150](https://github.com/kopexa-grc/kspec/issues/150)) ([05aae44](https://github.com/kopexa-grc/kspec/commit/05aae44052e81fc6e84c5161304fa6ae0d3a55a2))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatch ([#198](https://github.com/kopexa-grc/kspec/issues/198)) ([b7c4c1e](https://github.com/kopexa-grc/kspec/commit/b7c4c1e1978751a193d20f7e1c25e6b88581670d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatch ([#87](https://github.com/kopexa-grc/kspec/issues/87)) ([c91c66a](https://github.com/kopexa-grc/kspec/commit/c91c66a7bd8568f3bddd8950b8cbaa937b319b1b))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs ([#111](https://github.com/kopexa-grc/kspec/issues/111)) ([72e42a3](https://github.com/kopexa-grc/kspec/commit/72e42a3c82b6a4f13cdd6ee0b25047c61a27bae4))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs ([#129](https://github.com/kopexa-grc/kspec/issues/129)) ([c89adc5](https://github.com/kopexa-grc/kspec/commit/c89adc5e74551bb4b96c8fdfad55b1eeacbe760f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs ([#185](https://github.com/kopexa-grc/kspec/issues/185)) ([32a3463](https://github.com/kopexa-grc/kspec/commit/32a3463848c33115449a5f7f05de6afb4aa795ec))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/configservice ([#105](https://github.com/kopexa-grc/kspec/issues/105)) ([9aa0389](https://github.com/kopexa-grc/kspec/commit/9aa0389921d833047c61a863b09aa6091ed655b5))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/configservice ([#125](https://github.com/kopexa-grc/kspec/issues/125)) ([fba11a7](https://github.com/kopexa-grc/kspec/commit/fba11a75b9335b40a3cf81be64970025c89124e4))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/configservice ([#169](https://github.com/kopexa-grc/kspec/issues/169)) ([2121e2e](https://github.com/kopexa-grc/kspec/commit/2121e2e5fbb8023f3c1b34ff9e7651827746ac88))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/dynamodb ([#106](https://github.com/kopexa-grc/kspec/issues/106)) ([625a66e](https://github.com/kopexa-grc/kspec/commit/625a66e8397f51173809ddd4d77658e5ac8e5d07))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/dynamodb ([#124](https://github.com/kopexa-grc/kspec/issues/124)) ([2c8ca5e](https://github.com/kopexa-grc/kspec/commit/2c8ca5ec473dc4466954746c247d578ca95e575f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/dynamodb ([#158](https://github.com/kopexa-grc/kspec/issues/158)) ([84a3c9d](https://github.com/kopexa-grc/kspec/commit/84a3c9dc5c1ae8fe97f396044da1472f2630b4dd))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/dynamodb ([#180](https://github.com/kopexa-grc/kspec/issues/180)) ([60cb81a](https://github.com/kopexa-grc/kspec/commit/60cb81a57f5d4d03beb4bd1c9c44e5cdffdc54ab))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ec2 ([#127](https://github.com/kopexa-grc/kspec/issues/127)) ([5f1f400](https://github.com/kopexa-grc/kspec/commit/5f1f400f4f3229acf8852b3d8e9c52eede1fe13f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ec2 ([#159](https://github.com/kopexa-grc/kspec/issues/159)) ([077353e](https://github.com/kopexa-grc/kspec/commit/077353ef4029f1d1375b88acf22761561533354f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ec2 ([#177](https://github.com/kopexa-grc/kspec/issues/177)) ([f302734](https://github.com/kopexa-grc/kspec/commit/f3027344bf8ed0b3b4c253a7afda057c09ff5172))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ec2 ([#89](https://github.com/kopexa-grc/kspec/issues/89)) ([f4117d2](https://github.com/kopexa-grc/kspec/commit/f4117d2bff78b501197a9a65c53715621ffa4889))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ec2 ([#95](https://github.com/kopexa-grc/kspec/issues/95)) ([f1a75a6](https://github.com/kopexa-grc/kspec/commit/f1a75a6488c85b622f87fa6221705bc8d1312489))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecr ([#134](https://github.com/kopexa-grc/kspec/issues/134)) ([2f3fc5a](https://github.com/kopexa-grc/kspec/commit/2f3fc5a8e229b43adc92b8422f5cefcc0294feaf))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecr ([#192](https://github.com/kopexa-grc/kspec/issues/192)) ([6c579d9](https://github.com/kopexa-grc/kspec/commit/6c579d922efbb48308ec28ead7dc7d34baace04f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecr ([#94](https://github.com/kopexa-grc/kspec/issues/94)) ([c401119](https://github.com/kopexa-grc/kspec/commit/c4011199aa44e5f6aae2fb5c26d451f25947063e))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecs ([#123](https://github.com/kopexa-grc/kspec/issues/123)) ([ed82412](https://github.com/kopexa-grc/kspec/commit/ed82412e235062af78f1555b09a8421661972e9a))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecs ([#160](https://github.com/kopexa-grc/kspec/issues/160)) ([069d851](https://github.com/kopexa-grc/kspec/commit/069d8512e600a72f3357f1b3afd9e56f8ddd30ef))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecs ([#186](https://github.com/kopexa-grc/kspec/issues/186)) ([2be52d2](https://github.com/kopexa-grc/kspec/commit/2be52d264a759bde4f16a8c1d4f71de0b5cdc052))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecs ([#98](https://github.com/kopexa-grc/kspec/issues/98)) ([8f25070](https://github.com/kopexa-grc/kspec/commit/8f250702c917a26834fc532fc92b9abc9eb89f2f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/eks ([#116](https://github.com/kopexa-grc/kspec/issues/116)) ([cebf04a](https://github.com/kopexa-grc/kspec/commit/cebf04a6cb3cef2fb3f71992f360cb89fb4b566d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/eks ([#138](https://github.com/kopexa-grc/kspec/issues/138)) ([25624be](https://github.com/kopexa-grc/kspec/commit/25624be9d1ea9e1e14b48c7ca3cf8bbfedf82e6c))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/eks ([#173](https://github.com/kopexa-grc/kspec/issues/173)) ([586f4b2](https://github.com/kopexa-grc/kspec/commit/586f4b2e2d34b52364dcaa75e00eb6bd647be2cb))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/eks ([#88](https://github.com/kopexa-grc/kspec/issues/88)) ([33f2ce2](https://github.com/kopexa-grc/kspec/commit/33f2ce28e6dd93246f58cc62b11dbc18b829be1f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/elasticache ([#126](https://github.com/kopexa-grc/kspec/issues/126)) ([8e42bb7](https://github.com/kopexa-grc/kspec/commit/8e42bb782de6d29a57c1dadfbbe79ab6bf264ed0))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/elasticache ([#163](https://github.com/kopexa-grc/kspec/issues/163)) ([68ad4ce](https://github.com/kopexa-grc/kspec/commit/68ad4ce252d16419a84ce08fc2c7f857df607c7b))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 ([#113](https://github.com/kopexa-grc/kspec/issues/113)) ([46457d3](https://github.com/kopexa-grc/kspec/commit/46457d38853596b41d6f5e2ed1c90c0e68d2855b))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 ([#122](https://github.com/kopexa-grc/kspec/issues/122)) ([648ef77](https://github.com/kopexa-grc/kspec/commit/648ef77e1661566bc8c0b339db7001b8bf1004ad))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 ([#188](https://github.com/kopexa-grc/kspec/issues/188)) ([0a29365](https://github.com/kopexa-grc/kspec/commit/0a293656adbda13148f7cacb1ab8885e27547fc2))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/guardduty ([#142](https://github.com/kopexa-grc/kspec/issues/142)) ([56cac40](https://github.com/kopexa-grc/kspec/commit/56cac40e92218d4ea8be677fea05aa8cd3f47a46))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/guardduty ([#168](https://github.com/kopexa-grc/kspec/issues/168)) ([313f997](https://github.com/kopexa-grc/kspec/commit/313f9972e91508f7c69af265d163f3926b35b317))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/iam ([#109](https://github.com/kopexa-grc/kspec/issues/109)) ([b2893f8](https://github.com/kopexa-grc/kspec/commit/b2893f81d89bbdbe2365ccf3b8f428279c3a325c))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/iam ([#131](https://github.com/kopexa-grc/kspec/issues/131)) ([4fd0014](https://github.com/kopexa-grc/kspec/commit/4fd00140ccd8824dcae9d29f0e776261df4d54a1))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/iam ([#162](https://github.com/kopexa-grc/kspec/issues/162)) ([97f046c](https://github.com/kopexa-grc/kspec/commit/97f046c53551c46318fb5d9bc6816971275d21f0))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/kms ([#144](https://github.com/kopexa-grc/kspec/issues/144)) ([ebaf419](https://github.com/kopexa-grc/kspec/commit/ebaf419f4f6c328dc49d2f722f50a768509089a6))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/kms ([#179](https://github.com/kopexa-grc/kspec/issues/179)) ([1f08517](https://github.com/kopexa-grc/kspec/commit/1f085175afcc1c2d6931a2b9427a4b615bea9929))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/kms ([#99](https://github.com/kopexa-grc/kspec/issues/99)) ([433b4ca](https://github.com/kopexa-grc/kspec/commit/433b4cac72580f227e677c3bb3d6d251dcd8bf83))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/lambda ([#151](https://github.com/kopexa-grc/kspec/issues/151)) ([3174746](https://github.com/kopexa-grc/kspec/commit/317474625eb1fe2c8edaf169593cb94a65915712))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/lambda ([#196](https://github.com/kopexa-grc/kspec/issues/196)) ([bed3c0c](https://github.com/kopexa-grc/kspec/commit/bed3c0cd7ef033d787768e3c25884ce34c5bef09))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/organizations ([#114](https://github.com/kopexa-grc/kspec/issues/114)) ([97b6638](https://github.com/kopexa-grc/kspec/commit/97b663843560db0964a035de98fe510217af30b6))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/organizations ([#139](https://github.com/kopexa-grc/kspec/issues/139)) ([f321df7](https://github.com/kopexa-grc/kspec/commit/f321df78bde03b89eac966f99496c51cbc6002d8))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/organizations ([#174](https://github.com/kopexa-grc/kspec/issues/174)) ([68031b5](https://github.com/kopexa-grc/kspec/commit/68031b56b7dfbc1c98d923db284b18fbd65f044a))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/rds ([#115](https://github.com/kopexa-grc/kspec/issues/115)) ([d3e6ed9](https://github.com/kopexa-grc/kspec/commit/d3e6ed9d3d0b28f26eece738ba0377533eae1dfa))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/rds ([#137](https://github.com/kopexa-grc/kspec/issues/137)) ([c10642b](https://github.com/kopexa-grc/kspec/commit/c10642b0b30439f7b5e3b6e1a2ff934f6c330823))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/rds ([#187](https://github.com/kopexa-grc/kspec/issues/187)) ([e1411e3](https://github.com/kopexa-grc/kspec/commit/e1411e30841f836d1bac5c3e0240ed36c154cd33))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/rds ([#86](https://github.com/kopexa-grc/kspec/issues/86)) ([1237d04](https://github.com/kopexa-grc/kspec/commit/1237d0498a2f66b478014afd66fa7276cfdbcbb1))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/rds ([#96](https://github.com/kopexa-grc/kspec/issues/96)) ([0d02367](https://github.com/kopexa-grc/kspec/commit/0d023670b3b89405701fee19f3fad0e3da1bd3ff))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/s3 ([#133](https://github.com/kopexa-grc/kspec/issues/133)) ([1390b5f](https://github.com/kopexa-grc/kspec/commit/1390b5f48feac2f8f8c88dc5f79160045d27640e))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/s3 ([#172](https://github.com/kopexa-grc/kspec/issues/172)) ([4c1ca1f](https://github.com/kopexa-grc/kspec/commit/4c1ca1fe34989589e7ed985d55e7980572d5d1eb))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/secretsmanager ([#149](https://github.com/kopexa-grc/kspec/issues/149)) ([18aaae3](https://github.com/kopexa-grc/kspec/commit/18aaae3d3167e8ec05afe22baf3e3167e27592be))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/secretsmanager ([#191](https://github.com/kopexa-grc/kspec/issues/191)) ([2af7ca0](https://github.com/kopexa-grc/kspec/commit/2af7ca0b9ec417842bef499994c9ee6615a92b0b))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/securityhub ([#100](https://github.com/kopexa-grc/kspec/issues/100)) ([38103e8](https://github.com/kopexa-grc/kspec/commit/38103e86d7b2d58c8f75890b284f647523367040))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/securityhub ([#118](https://github.com/kopexa-grc/kspec/issues/118)) ([ef56ab3](https://github.com/kopexa-grc/kspec/commit/ef56ab3f00ca613430be5b07aae6b931a0f077a0))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/securityhub ([#143](https://github.com/kopexa-grc/kspec/issues/143)) ([eafddf2](https://github.com/kopexa-grc/kspec/commit/eafddf2248e88498e596ab7740c33e3583bd1b0f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/securityhub ([#170](https://github.com/kopexa-grc/kspec/issues/170)) ([77ea181](https://github.com/kopexa-grc/kspec/commit/77ea181d4661bc7186e06ce8205f81203817fd24))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/sns ([#107](https://github.com/kopexa-grc/kspec/issues/107)) ([a8c403c](https://github.com/kopexa-grc/kspec/commit/a8c403ccbf656049400331af7ee0f3a6016218a3))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/sns ([#152](https://github.com/kopexa-grc/kspec/issues/152)) ([4ed5459](https://github.com/kopexa-grc/kspec/commit/4ed54596b82dc6bdcc335eaa530705621dc36df5))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/sns ([#184](https://github.com/kopexa-grc/kspec/issues/184)) ([b83bae9](https://github.com/kopexa-grc/kspec/commit/b83bae9b5c654e6c8fe59519954ca6d38782302d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/sqs ([#132](https://github.com/kopexa-grc/kspec/issues/132)) ([0f00be8](https://github.com/kopexa-grc/kspec/commit/0f00be881ed6504959b01ea1025e6aacf800482f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/sqs ([#176](https://github.com/kopexa-grc/kspec/issues/176)) ([998b2c7](https://github.com/kopexa-grc/kspec/commit/998b2c7a0d6ff0ba36b2e781a6ce43c6770e8b9d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ssm ([#141](https://github.com/kopexa-grc/kspec/issues/141)) ([9627672](https://github.com/kopexa-grc/kspec/commit/9627672cc8f03c20a223a1ecd5b231990829614f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ssm ([#164](https://github.com/kopexa-grc/kspec/issues/164)) ([2ede6e5](https://github.com/kopexa-grc/kspec/commit/2ede6e5d678dffd2e189e3b763753a300a529899))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ssm ([#92](https://github.com/kopexa-grc/kspec/issues/92)) ([224ecb4](https://github.com/kopexa-grc/kspec/commit/224ecb411459461a4d1643158568e390354973eb))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/wafv2 ([#154](https://github.com/kopexa-grc/kspec/issues/154)) ([569da2f](https://github.com/kopexa-grc/kspec/commit/569da2fc90cfb92ea2d8a2526ecfaece90de5173))
* **deps:** bump github.com/charmbracelet/bubbles from 0.21.0 to 1.0.0 ([#101](https://github.com/kopexa-grc/kspec/issues/101)) ([b6b4c75](https://github.com/kopexa-grc/kspec/commit/b6b4c754006467f3220db65f5a32d9fe4014471d))
* **deps:** bump github.com/hetznercloud/hcloud-go/v2 ([#189](https://github.com/kopexa-grc/kspec/issues/189)) ([3c6d6dd](https://github.com/kopexa-grc/kspec/commit/3c6d6dddb5e8928684937fb2d622f85affb63232))
* **deps:** bump github.com/microsoftgraph/msgraph-sdk-go ([#91](https://github.com/kopexa-grc/kspec/issues/91)) ([ca6c65f](https://github.com/kopexa-grc/kspec/commit/ca6c65fa80698275537b43e0eb0ca96e6cacfb08))
* **deps:** bump github.com/rs/zerolog from 1.34.0 to 1.35.0 ([#167](https://github.com/kopexa-grc/kspec/issues/167)) ([4bb0129](https://github.com/kopexa-grc/kspec/commit/4bb0129cfa574c074d5a73ae1e661445665470d2))
* **deps:** bump github.com/xuri/excelize/v2 from 2.10.0 to 2.10.1 ([#104](https://github.com/kopexa-grc/kspec/issues/104)) ([33da0ab](https://github.com/kopexa-grc/kspec/commit/33da0ab19536f01c1678b28e18a8b1720e7d6d9f))
* **deps:** bump github.com/yuin/goldmark from 1.7.16 to 1.7.17 ([#157](https://github.com/kopexa-grc/kspec/issues/157)) ([f99e198](https://github.com/kopexa-grc/kspec/commit/f99e198284dda8e6dd79625a7b6eb3493e7e0806))
* **deps:** bump golang.org/x/oauth2 from 0.35.0 to 0.36.0 ([#130](https://github.com/kopexa-grc/kspec/issues/130)) ([3cce66b](https://github.com/kopexa-grc/kspec/commit/3cce66bb8ea67dc73b772b0a4ef977a66433a8c5))
* **deps:** bump golang.org/x/time from 0.14.0 to 0.15.0 ([#135](https://github.com/kopexa-grc/kspec/issues/135)) ([45a8519](https://github.com/kopexa-grc/kspec/commit/45a85195d36cd0dc767582d49a24cfe0c44e1693))

## [0.2.3](https://github.com/kopexa-grc/kspec/compare/v0.2.2...v0.2.3) (2026-02-11)


### Features

* **examples:** add per-asset disabled queries to worker example ([2f022dd](https://github.com/kopexa-grc/kspec/commit/2f022dd5f722e041ebf245c8846f5b962f1e1c57))


### Documentation

* **examples:** add high-availability worker example ([e57a982](https://github.com/kopexa-grc/kspec/commit/e57a9828e86f376564502fbb04d4a3318f058611))
* **integration:** add integration guide and working examples ([ce6450c](https://github.com/kopexa-grc/kspec/commit/ce6450c39ff534a61ba26c20d4473b18b056bdf6))


### Dependencies

* **deps:** bump github.com/aws/aws-sdk-go-v2/service/autoscaling ([#69](https://github.com/kopexa-grc/kspec/issues/69)) ([0f4a1c2](https://github.com/kopexa-grc/kspec/commit/0f4a1c26318353e1f6db9e62ddcfdde8e3cb36f6))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudfront ([#84](https://github.com/kopexa-grc/kspec/issues/84)) ([7896c36](https://github.com/kopexa-grc/kspec/commit/7896c36010d4c4837c6a7c49da74bb19f9bed3e8))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/configservice ([#68](https://github.com/kopexa-grc/kspec/issues/68)) ([7be85df](https://github.com/kopexa-grc/kspec/commit/7be85dfb7f21b09cd1492bb850c7f3cbec23662b))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/dynamodb ([#73](https://github.com/kopexa-grc/kspec/issues/73)) ([5443094](https://github.com/kopexa-grc/kspec/commit/5443094330d961a01a102c03ceeb20a0b3531992))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/dynamodb ([#80](https://github.com/kopexa-grc/kspec/issues/80)) ([46f2b85](https://github.com/kopexa-grc/kspec/commit/46f2b8598a75de95fbdf9bb81900de8d5ee2f050))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ec2 ([#70](https://github.com/kopexa-grc/kspec/issues/70)) ([46337e5](https://github.com/kopexa-grc/kspec/commit/46337e5be9a03790cfbc17e181dbe275924ebd59))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ec2 ([#76](https://github.com/kopexa-grc/kspec/issues/76)) ([28d1ff8](https://github.com/kopexa-grc/kspec/commit/28d1ff844bb5232841aa9b710deba46ef3fc61f8))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/eks ([#83](https://github.com/kopexa-grc/kspec/issues/83)) ([34b2c9f](https://github.com/kopexa-grc/kspec/commit/34b2c9fe70a4e9e7f8a8484dc3335cd939a248b8))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/guardduty ([#71](https://github.com/kopexa-grc/kspec/issues/71)) ([40b3528](https://github.com/kopexa-grc/kspec/commit/40b35283e574fc2133aa55b3d3a3a49cb9d5837e))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/lambda ([#77](https://github.com/kopexa-grc/kspec/issues/77)) ([db21df3](https://github.com/kopexa-grc/kspec/commit/db21df36fa9cbed491b4da07b282b5ed651a27a5))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/organizations ([#81](https://github.com/kopexa-grc/kspec/issues/81)) ([bd30422](https://github.com/kopexa-grc/kspec/commit/bd30422bb496fe3a4bd7acb15483666c230f376f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/s3 ([#75](https://github.com/kopexa-grc/kspec/issues/75)) ([de66638](https://github.com/kopexa-grc/kspec/commit/de66638d4755174d305416f4a86d218e682cdc0d))
* **deps:** bump github.com/google/cel-go in the google group ([#74](https://github.com/kopexa-grc/kspec/issues/74)) ([4b7cf1b](https://github.com/kopexa-grc/kspec/commit/4b7cf1b0ce77ce6483850ef9bd3e22069666a141))
* **deps:** bump github.com/hetznercloud/hcloud-go/v2 ([#72](https://github.com/kopexa-grc/kspec/issues/72)) ([7be3666](https://github.com/kopexa-grc/kspec/commit/7be3666cebd7a05782ad8861bace904abd34b302))
* **deps:** bump github.com/microsoftgraph/msgraph-sdk-go ([#66](https://github.com/kopexa-grc/kspec/issues/66)) ([52adeaf](https://github.com/kopexa-grc/kspec/commit/52adeafd37f554290bdb6c13a974c5b2e9ffab42))
* **deps:** bump github.com/microsoftgraph/msgraph-sdk-go ([#78](https://github.com/kopexa-grc/kspec/issues/78)) ([ae55381](https://github.com/kopexa-grc/kspec/commit/ae55381f9fbd17e7d9f74cdc00307bac3aa03395))
* **deps:** bump github.com/miekg/dns from 1.1.70 to 1.1.72 ([#67](https://github.com/kopexa-grc/kspec/issues/67)) ([2017802](https://github.com/kopexa-grc/kspec/commit/20178024682ddf326da5186a3438a54ea87fb8bd))
* **deps:** bump golang.org/x/oauth2 from 0.34.0 to 0.35.0 ([#82](https://github.com/kopexa-grc/kspec/issues/82)) ([4e806b4](https://github.com/kopexa-grc/kspec/commit/4e806b4c954bf0edd5186f1eded3dfe99d1652ab))

## [0.2.2](https://github.com/kopexa-grc/kspec/compare/v0.2.1...v0.2.2) (2026-01-19)


### Features

* **scoring:** add graph-based scoring system with policy-driven configuration ([#65](https://github.com/kopexa-grc/kspec/issues/65)) ([c8bb6f6](https://github.com/kopexa-grc/kspec/commit/c8bb6f638ca0251ef284fd9d1310ae7493a21b80))


### Code Refactoring

* consolidate provider registry and move CEL evaluator ([#63](https://github.com/kopexa-grc/kspec/issues/63)) ([4173ab1](https://github.com/kopexa-grc/kspec/commit/4173ab198fbdbf21ef92f1b7fdf32d6fb089bdad))

## [0.2.1](https://github.com/kopexa-grc/kspec/compare/v0.2.0...v0.2.1) (2026-01-19)


### Features

* add discovery command and graph-based resource traversal ([#51](https://github.com/kopexa-grc/kspec/issues/51)) ([c25768a](https://github.com/kopexa-grc/kspec/commit/c25768a592f451a61edca17f1091f55ece4b5663))


### Documentation

* **readme:** update commands to use asset type subcommands ([b20d9e2](https://github.com/kopexa-grc/kspec/commit/b20d9e2602a32d5024472e1624b834a63d7b91be))


### Dependencies

* **deps:** bump github.com/aws/aws-sdk-go-v2/service/autoscaling ([#56](https://github.com/kopexa-grc/kspec/issues/56)) ([2afba51](https://github.com/kopexa-grc/kspec/commit/2afba5176bb91889276ccf48e0f5dd064c8713e2))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudtrail ([#58](https://github.com/kopexa-grc/kspec/issues/58)) ([5787049](https://github.com/kopexa-grc/kspec/commit/5787049b98b0a75a43c5224e1257769184b699c6))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ec2 ([#59](https://github.com/kopexa-grc/kspec/issues/59)) ([517a3ca](https://github.com/kopexa-grc/kspec/commit/517a3cade01baa649c217bb22184083e57b3ba02))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecs ([#60](https://github.com/kopexa-grc/kspec/issues/60)) ([97e3bf1](https://github.com/kopexa-grc/kspec/commit/97e3bf1ff7eee1dad7a7c66c8beeaeee841dd497))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/secretsmanager ([#55](https://github.com/kopexa-grc/kspec/issues/55)) ([f539f5c](https://github.com/kopexa-grc/kspec/commit/f539f5c3d163171d2d61d860f1ce0e84daac1f84))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/sns ([#61](https://github.com/kopexa-grc/kspec/issues/61)) ([5b305be](https://github.com/kopexa-grc/kspec/commit/5b305be058c55bcdb5ff68566e8bc7d7d696797f))
* **deps:** bump github.com/Azure/azure-sdk-for-go/sdk/azcore ([#54](https://github.com/kopexa-grc/kspec/issues/54)) ([c72851d](https://github.com/kopexa-grc/kspec/commit/c72851d745bcba09025ed75ddc2b6f1c681f04d8))
* **deps:** bump github.com/hetznercloud/hcloud-go/v2 ([#57](https://github.com/kopexa-grc/kspec/issues/57)) ([a4d07b0](https://github.com/kopexa-grc/kspec/commit/a4d07b01c6602a6441358b98b7d4ab2e97cb8926))

## [0.2.0](https://github.com/kopexa-grc/kspec/compare/v0.1.6...v0.2.0) (2026-01-18)


### ⚠ BREAKING CHANGES

* **provider:** Default policy directory

### Features

* concurrent scanning, provider refactoring, and expanded AWS security policies ([#50](https://github.com/kopexa-grc/kspec/issues/50)) ([8907e4a](https://github.com/kopexa-grc/kspec/commit/8907e4a39f36444b90646af2b1f73c77e11eab15))
* **provider:** implement dynamic self-registration pattern ([7ad2807](https://github.com/kopexa-grc/kspec/commit/7ad2807ce592f770ed35c5bcd8d8be262163dde7)), closes [#47](https://github.com/kopexa-grc/kspec/issues/47)
* **report:** add HTML export format ([#48](https://github.com/kopexa-grc/kspec/issues/48)) ([d1cea21](https://github.com/kopexa-grc/kspec/commit/d1cea21422a462d90c0db364b867acdd45fbcbf4))
* **report:** add HTML export format with interactive features ([d1cea21](https://github.com/kopexa-grc/kspec/commit/d1cea21422a462d90c0db364b867acdd45fbcbf4))
* **report:** add report export and non-interactive scan mode ([19aa52d](https://github.com/kopexa-grc/kspec/commit/19aa52d71ffe6825fb9e67c99ee7103075c71770))


### Dependencies

* **deps:** bump github.com/aws/aws-sdk-go-v2/config ([#30](https://github.com/kopexa-grc/kspec/issues/30)) ([af4856b](https://github.com/kopexa-grc/kspec/commit/af4856b9eb1fc428f605d1c7cb8c1a3567bee742))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/acm ([#37](https://github.com/kopexa-grc/kspec/issues/37)) ([8c7a397](https://github.com/kopexa-grc/kspec/commit/8c7a3976fabf1959c20585eace85b2b8b313e78c))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/apigateway ([#27](https://github.com/kopexa-grc/kspec/issues/27)) ([22f10e0](https://github.com/kopexa-grc/kspec/commit/22f10e0a85c8438d715ada08049872d5083c480a))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/apigatewayv2 ([#22](https://github.com/kopexa-grc/kspec/issues/22)) ([a2694b5](https://github.com/kopexa-grc/kspec/commit/a2694b5310fa48abf5198c37d1ca6da8516e9c70))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudfront ([#38](https://github.com/kopexa-grc/kspec/issues/38)) ([8fd18af](https://github.com/kopexa-grc/kspec/commit/8fd18af5b4bba81f53c1df181554ab83ac0e3147))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatch ([#39](https://github.com/kopexa-grc/kspec/issues/39)) ([24b0fd8](https://github.com/kopexa-grc/kspec/commit/24b0fd86a2e20bc484bb875b1c942085dde3cc4c))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs ([#19](https://github.com/kopexa-grc/kspec/issues/19)) ([7f89f5f](https://github.com/kopexa-grc/kspec/commit/7f89f5fb38b9eec918c412450a7152d39d5cb24a))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/configservice ([#44](https://github.com/kopexa-grc/kspec/issues/44)) ([bba331c](https://github.com/kopexa-grc/kspec/commit/bba331ca74c09ef78ba8e1d0763baab90c2b2607))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/dynamodb ([#45](https://github.com/kopexa-grc/kspec/issues/45)) ([40bdf9b](https://github.com/kopexa-grc/kspec/commit/40bdf9bb09ea3580e915e1560b692ef8ce5b007d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecr ([#23](https://github.com/kopexa-grc/kspec/issues/23)) ([9c38402](https://github.com/kopexa-grc/kspec/commit/9c384022924f1e59887075edd8320667b58423a6))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/eks ([#34](https://github.com/kopexa-grc/kspec/issues/34)) ([be480fa](https://github.com/kopexa-grc/kspec/commit/be480fafda7a591474bd9db7851107ec184c37d3))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/elasticache ([#20](https://github.com/kopexa-grc/kspec/issues/20)) ([449bd22](https://github.com/kopexa-grc/kspec/commit/449bd228dafb5643383c5eec63d3cff6f9b4ff5d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 ([#41](https://github.com/kopexa-grc/kspec/issues/41)) ([3a30a91](https://github.com/kopexa-grc/kspec/commit/3a30a916d91fd082414cf390e1e5f0d1445177d3))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/guardduty ([#33](https://github.com/kopexa-grc/kspec/issues/33)) ([8a1cb14](https://github.com/kopexa-grc/kspec/commit/8a1cb1412abc3e83bbea17d46892f0dfaef072ea))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/iam ([#43](https://github.com/kopexa-grc/kspec/issues/43)) ([ad6bd26](https://github.com/kopexa-grc/kspec/commit/ad6bd263e65a9bc106feb50580714f9d17fef96f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/kms ([#42](https://github.com/kopexa-grc/kspec/issues/42)) ([c76998f](https://github.com/kopexa-grc/kspec/commit/c76998ffd53937601297036a78235b36719f0cb6))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/lambda ([#21](https://github.com/kopexa-grc/kspec/issues/21)) ([ae004e6](https://github.com/kopexa-grc/kspec/commit/ae004e6a026803c436dca7dba35de20a20199aa1))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/organizations ([#32](https://github.com/kopexa-grc/kspec/issues/32)) ([01b6e7b](https://github.com/kopexa-grc/kspec/commit/01b6e7b46b8cddc65c316d22ab19520558d1bd3c))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/rds ([#40](https://github.com/kopexa-grc/kspec/issues/40)) ([5d4c56e](https://github.com/kopexa-grc/kspec/commit/5d4c56e46d2a911d5032553c3325454ae627f896))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/s3 ([#26](https://github.com/kopexa-grc/kspec/issues/26)) ([ee03308](https://github.com/kopexa-grc/kspec/commit/ee033084d23ce4cf00e65dda9ff0132bc00ffc10))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/securityhub ([#36](https://github.com/kopexa-grc/kspec/issues/36)) ([896822a](https://github.com/kopexa-grc/kspec/commit/896822a10724bfe542d747e685e5241d792e4802))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/sqs ([#31](https://github.com/kopexa-grc/kspec/issues/31)) ([dbfac6d](https://github.com/kopexa-grc/kspec/commit/dbfac6d2550fd11cadce8380e1bcff0f8d5fd71b))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ssm ([#24](https://github.com/kopexa-grc/kspec/issues/24)) ([9c9a446](https://github.com/kopexa-grc/kspec/commit/9c9a446affbe527db995d320dcc73a95375b880d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/wafv2 ([#25](https://github.com/kopexa-grc/kspec/issues/25)) ([406f91b](https://github.com/kopexa-grc/kspec/commit/406f91bbfc5d0c17d4c96be6dd4124f92dbd82b6))
* **deps:** bump github.com/hetznercloud/hcloud-go/v2 ([#28](https://github.com/kopexa-grc/kspec/issues/28)) ([35d2eb0](https://github.com/kopexa-grc/kspec/commit/35d2eb009eaeca538155dadf98be7f7105c53164))
* **deps:** bump github.com/microsoftgraph/msgraph-sdk-go ([#14](https://github.com/kopexa-grc/kspec/issues/14)) ([088ba75](https://github.com/kopexa-grc/kspec/commit/088ba75a0be7d981132237791604019f0706edb8))
* **deps:** bump github.com/miekg/dns from 1.1.69 to 1.1.70 ([#29](https://github.com/kopexa-grc/kspec/issues/29)) ([0e435f4](https://github.com/kopexa-grc/kspec/commit/0e435f436fc7b26fc7d1dc299e3168fabab79f37))

## [0.1.6](https://github.com/kopexa-grc/kspec/compare/v0.1.5...v0.1.6) (2026-01-08)


### Features

* add ptr package and optimize AWS provider with fuzz tests ([7ebe687](https://github.com/kopexa-grc/kspec/commit/7ebe6879196be67c0d5492815d714fc1834a8b42))
* **azure:** add comprehensive Azure provider resources ([c8a3216](https://github.com/kopexa-grc/kspec/commit/c8a32162cd8927a7eb8c61b237533678af1d77e9))
* **schema:** add JSON Schema generation for policy validation ([d0f59b9](https://github.com/kopexa-grc/kspec/commit/d0f59b9278ca74de0cda020c542f34cf67619bc8))


### Bug Fixes

* preallocate slices to satisfy prealloc linter ([a90f8af](https://github.com/kopexa-grc/kspec/commit/a90f8afbb4ffa60e8b9b269f76dc573c7e1072c0))
* resolve remaining prealloc linter issues for golangci-lint v2.8.0 ([3b2efd9](https://github.com/kopexa-grc/kspec/commit/3b2efd9d82dcb23354d18ecd2f283decf933532d))
* update license to ELv2 in policy files ([41b7593](https://github.com/kopexa-grc/kspec/commit/41b75934bf83d85be088d7a6a47a04f77eb4119e))


### Code Refactoring

* add core helpers and improve error handling ([ea357ff](https://github.com/kopexa-grc/kspec/commit/ea357ffc014911e68628272f98c22945653ea12f))

## [0.1.5](https://github.com/kopexa-grc/kspec/compare/v0.1.4...v0.1.5) (2026-01-07)


### Features

* **aws:** add comprehensive AWS provider for security scanning ([51e4383](https://github.com/kopexa-grc/kspec/commit/51e4383a9b87eeda52863c3340d09e14c1d3a07b))


### Bug Fixes

* resolve linter issues and add provider documentation ([ae67317](https://github.com/kopexa-grc/kspec/commit/ae67317f36f08a185d6ecff52b3714181f9dc6ee))

## [0.1.4](https://github.com/kopexa-grc/kspec/compare/v0.1.3...v0.1.4) (2026-01-06)


### Features

* add enterprise security features ([e53fc94](https://github.com/kopexa-grc/kspec/commit/e53fc949d50a1fed53810b178423e8e004f19992))
* add signed releases and SLSA provenance ([cd290a9](https://github.com/kopexa-grc/kspec/commit/cd290a9714385b6f503ebcfe6a652d4762079306))
* **factorial:** add Factorial HR provider for compliance scanning ([31972b9](https://github.com/kopexa-grc/kspec/commit/31972b99edf3f5c210ea039cec8ec80b2f80287e))


### Bug Fixes

* resolve linter issues and add test coverage ([4b986db](https://github.com/kopexa-grc/kspec/commit/4b986db0cf5ac07c0023dc379ce3f1221b9c774e))
* standardize policy YAML files to match Go struct ([b36b00e](https://github.com/kopexa-grc/kspec/commit/b36b00e546542ea1b858629770fac5d6ffc2cff1))


### Documentation

* add comprehensive provider documentation ([9ed122f](https://github.com/kopexa-grc/kspec/commit/9ed122fc3b356c0bbce1e66111a5494e898f3817))
* reorganize and complete documentation structure ([1287a4e](https://github.com/kopexa-grc/kspec/commit/1287a4e990dbd4fea8482e4b7fed4020591140de))

## [0.1.3](https://github.com/kopexa-grc/kspec/compare/v0.1.2...v0.1.3) (2026-01-06)


### Bug Fixes

* **ci:** fix archive upload and checksum generation ([1c252f1](https://github.com/kopexa-grc/kspec/commit/1c252f159f2cfb36332063d7989ef12510dd789b))

## [0.1.2](https://github.com/kopexa-grc/kspec/compare/v0.1.1...v0.1.2) (2026-01-06)


### Bug Fixes

* use default font in demo GIF to fix letter spacing ([27dd58d](https://github.com/kopexa-grc/kspec/commit/27dd58dd193217ab250dd17f0ff286a162f70131))


### Documentation

* add demo GIF showcasing host scanning ([016cd41](https://github.com/kopexa-grc/kspec/commit/016cd41ebcdfcc3676528f3f8d57fce308ff27d9))

## [0.1.1](https://github.com/kopexa-grc/kspec/compare/v0.1.0...v0.1.1) (2026-01-06)


### Features

* Add contribution guidelines, issue templates, and security policy ([4366e66](https://github.com/kopexa-grc/kspec/commit/4366e660e872934f2af8203b4a52ee96ee8b53e5))
* Add GitHub Actions workflow for automated releases using Release Please ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))
* Add GitHub organization and repository scanning with flexible credential handling. ([07359af](https://github.com/kopexa-grc/kspec/commit/07359af686e2d0ecda1504fea665b4810728d6fc))
* Add Makefile, golangci-lint configuration, and Lefthook setup for improved development workflow ([6fa8896](https://github.com/kopexa-grc/kspec/commit/6fa8896e4b06b4a9405ac456792acbd83698b494))
* Add SBOM component and vulnerability resources with tests ([1dfecfb](https://github.com/kopexa-grc/kspec/commit/1dfecfb2e337ac0761610b0d234456cab02d76c2))
* **atlassian:** add Jira security scheme and user resources ([79ed2ac](https://github.com/kopexa-grc/kspec/commit/79ed2acf2e079569730e793acedcde8628e4d02f))
* azure ([c8dcd76](https://github.com/kopexa-grc/kspec/commit/c8dcd762f21fe671194c487072f815a7f769a50c))
* cli ([42d5d63](https://github.com/kopexa-grc/kspec/commit/42d5d63c09d18a6959489c4c45b9ac1e0eaf3a89))
* **cloudflare:** add support for various Cloudflare resources ([7518757](https://github.com/kopexa-grc/kspec/commit/75187578f795a055435a26eea76dd3c5a1adca28))
* Create GoReleaser configuration for building and releasing kspec ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))
* enhanced TUI, certificate scanning, and quickstart docs ([acf87f1](https://github.com/kopexa-grc/kspec/commit/acf87f1ad679629216ba096cfbc562f6092b93eb))
* first version ([2da72ea](https://github.com/kopexa-grc/kspec/commit/2da72eac7c700be9fd07a4ea361924dcf51b49c7))
* **hetzner:** add Hetzner Cloud provider for infrastructure security scanning ([b142ceb](https://github.com/kopexa-grc/kspec/commit/b142cebb2f510faf0d24e1c151b3e2b5f304cc8b))
* introduce CLI with scan command for policy evaluation and TUI results ([4b25063](https://github.com/kopexa-grc/kspec/commit/4b25063cf717f1b0779e3e63f18728fe00785ae3))
* **ms365:** Add Microsoft 365 provider with Teams and Tenant resources ([4fb91b2](https://github.com/kopexa-grc/kspec/commit/4fb91b298519873a88bf84b84e20cda832e56bab))
* remove `cnquery` command and `example_policy.yaml`, update `README.md` for `kspec` command, and add `kspec` to `.gitignore` ([8214b0e](https://github.com/kopexa-grc/kspec/commit/8214b0ea8bcf2fdf9a0f261177a74c9a4bf03284))


### Bug Fixes

* address linting issues across the codebase ([b28da0c](https://github.com/kopexa-grc/kspec/commit/b28da0c90e055a5e06af0e792f61bc9cb0d00ef3))
* resolve all remaining linter issues ([31ad096](https://github.com/kopexa-grc/kspec/commit/31ad0965aa0be99a4ac6c7690c279023c07b5b57))
* resolve revive linter stuttering and package-comments issues ([c88351b](https://github.com/kopexa-grc/kspec/commit/c88351bcc2d87d5315d8698545ee7066cf11040a))
* Update Azure provider imports to new repository path ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))
* Update golangci-lint configuration and change changelog type to default ([8754798](https://github.com/kopexa-grc/kspec/commit/8754798a47b3e8fd54aea63f5215ca49c0aed500))
* Update import paths to reflect new repository structure ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))
* Update README.md with new repository link and author information ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))


### Documentation

* Create CHANGELOG.md to document notable changes ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))
* Update project banner image. ([bec4729](https://github.com/kopexa-grc/kspec/commit/bec4729b9fc61fc094c868570887c7ce8182c7a4))


### Code Refactoring

* Clean up comments and improve clarity in policy files; add git log permission ([e218ea9](https://github.com/kopexa-grc/kspec/commit/e218ea9662d752596e05105154948b82518e5c03))
* Update all provider imports to point to the new repository path ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))

## [Unreleased]

### Features

- Initial release of kspec policy-as-code engine
- Azure provider with support for storage accounts, SQL servers, Key Vaults, NSGs, VMs, and App Services
- Microsoft 365 provider with support for users, groups, applications, Teams, security policies, and more
- GitHub provider with support for organizations, repositories, branches, and teams
- Network provider for TLS, DNS, and HTTP security scanning
- CEL-based policy evaluation engine
- Interactive TUI for real-time scan progress
- Policy-as-code YAML format with comprehensive documentation

### Documentation

- Azure provider setup guide
- Microsoft 365 provider setup guide
- GitHub provider setup guide
- Example security policies for all providers
