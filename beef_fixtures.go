package bux

// Fixtures for beef_tx_test.go
var expectedBeefHex = map[int]string{
	1: "0100beef02fea6790c000f02fda82c000703145efd05fec7c2edc1827ec685755a2d05208486e9bc2461268e0f5e533bfda92c02cb3553424ffc94b59a60fb358b6cb6dfb694aee894dcd1effc0ed0a9052464e301fd551600b7b53a09331453b5966589dc473b45c87823f109417593a82ff8ffe7ddc6f96e01fd2b0b0080085c6f18f35d5f0a231eda36c1be3f031734bb6cb6987978ef2aad007ea6da01fd940500d26d24b44097ed3f0c2927413dd2f1fb78bba948803abdb7f2fb51d9807a77bc01fdcb0200bad82dfd55455709713ea8390a7c76be5c076da9cd487b75f558728ef8572b7601fd64010022553e159788764f6b3c1e27324f99a13abee9c7069ce34b8a4fcabc45b7aabb01b300debfa516f54f4331ddf74067403d8a70915973e8300c298ff8dbe5bdcc94768101580079a933dbaeaee5dfb6af1ce9bd7a9ef40584e1844a938e398d19f94aba525bed012d005287b1c986e6495d00c103082618d9e0a30c0c620fd2d232eb9c31d59dec510e01170074639f74ebdaaac679e8fed504ae62a6f42c633a549a99e21940aff14e69ab08010a008eb3fde752d9e5c8e67c26daa95f8b9480ec11e608de7afae04eb19775f42342010400eea8e24927e93a0cf9ca50b315e48efca1f0b77643f7110c6efd94220dcedab6010300682c5f78d5f89e5e5d13619f543aacb00d1e4e85043f4d453bd6f6eb14755e790100006df75ca701801279bcb9b91579ac74b2131913ba35a0c3cb313c49062c8f453f0101009fe83bf3febfc2af4641b9bc8c1f76ff9b2ae7de9f4ab5374e672ac70cf8b9b9fec27f0c000b02fd8802000e40280d05fedfd66af3edb59a94f0512e804093f855c025b9d964ca23ca1b12fd890202624fbcb4e68d162361f456b8b4fef6b9e7943013088b32b6bca7f5ced41ff00401fd45010090f9751ef8c4daa8a15d0cc75d1a89e7fad8b8ba258f36da56fbb8faef725b1e01a3002863542b3fe0a1a8fdd6711f7b9af6d9fa2986bf6c67460e5e7fc544a3fa7ec90150003a0b1e497aae08b126ec790f9214517e32cff1a57cf3abe92e9580f634f369500129006d6f372a6b54acb13e4c0ce2b6ff53120685ac23bf8e1953f8358b87ebf90ca4011500c310dbdea5f87b557964e09c67926b6d756c9ea821948e36298544bf761bf9b5010b00a267e36fe2eef4ce3582e359d6b0b93df7d76afd00e19d4abff77d7cf2acf6db01040080c16fd797c5b4f8463e5bffc6ffacd7dae409c9908b4748ff28dabd8bbe8fe0010300b7e09ca039e8b8052f179d5f2c3d1d6e98bb983470784d62f03438621effe80a010000cf5b6da719ca8b752c50b5b9b4d7659cc5aaed7fe429d96d2270617414fb551f010100dc01860ec79aac9c4465b6afb0b0641bbf0b1c8a1c23bf7b920a1b662366ab40030100000001a114c7deb8deba851d87755aa10aa18c97bd77afee4e1bad01d1c50e07a644eb010000006a473044022041abd4f93bd1db1d0097f2d467ae183801d7842d23d0605fa9568040d245167402201be66c96bef4d6d051304f6df2aecbdfe23a8a05af0908ef2117ab5388d8903c412103c08545a40c819f6e50892e31e792d221b6df6da96ebdba9b6fe39305cc6cc768ffffffff0263040000000000001976a91454097d9d921f9a1f55084a943571d868552e924f88acb22a0000000000001976a914c36b3fca5159231033f3fbdca1cde942096d379f88ac0000000001010100000001cfc39e3adcd58ed58cf590079dc61c3eb6ec739abb7d22b592fb969d427f33ee000000006a4730440220253e674e64028459457d55b444f5f3dc15c658425e3184c628016739e4921fd502207c8fe20eb34e55e4115fbd82c23878b4e54f01f6c6ad0811282dd0b1df863b5e41210310a4366fd997127ad972b14c56ca2e18f39ca631ac9e3e4ad3d9827865d0cc70ffffffff0264000000000000001976a914668a92ff9cb5785eb8fc044771837a0818b028b588acdc4e0000000000001976a914b073264927a61cf84327dea77414df6c28b11e5988ac0000000001000100000002cb3553424ffc94b59a60fb358b6cb6dfb694aee894dcd1effc0ed0a9052464e3000000006a4730440220515c3bf93d38fa7cc164746fae4bec8b66c60a82509eb553751afa5971c3e41d0220321517fd5c997ab5f8ef0e59048ce9157de46f92b10d882bf898e62f3ee7343d4121038f1273fcb299405d8d140b4de9a2111ecb39291b2846660ebecd864d13bee575ffffffff624fbcb4e68d162361f456b8b4fef6b9e7943013088b32b6bca7f5ced41ff004010000006a47304402203fb24f6e00a6487cf88a3b39d8454786db63d649142ea76374c2f55990777e6302207fbb903d038cf43e13ffb496a64f36637ec7323e5ac48bb96bdb4a885100abca4121024b003d3cf49a8f48c1fe79b711b1d08e306c42a0ab8da004d97fccc4ced3343affffffff026f000000000000001976a914f232d38cd4c2f87c117af06542b04a7061b6640188aca62a0000000000001976a9146058e52d00e3b94211939f68cc2d9a3fc1e3db0f88ac0000000000",
	2: "0100beef01fe41800c000b06fd2c0402e01e74cab9a0571ab5a7d86794826f756a9c65dd0dea3bb3720c4051c488cf50fd2d0400d9060c543afb1c0faafb96667ed788324d4d1c338142a0841fe3ab9c30922cb4fd90040208c461a39a8877db46472f5cc59e5a108e417b1c9ea3091b71b65346d218f471fd910400ff8fa1e395088748feae2d7729ab9d5da0225f5ed80b2f295625f7c77da087f4fdcc05022256c94d07451664749e440f55cec8a37da1c46cf30a97579e2f9696b84ad484fdcd0500aecbe0a519d483bad8758c3a69cdc0dc12b19a363fded579ebc993edf510746503fd170200dd5d795e63f8777ef7a82453946150946435e52d4076089ce0cb15d8a1237c84fd4902000f3311a938e3f7977bb7a2db5ca912e4e0f26bd12744051333cd22cd3a2fad89fde702008695a1dfeec9393365a21690b089018c9d7dd94bbbf85b62f48701424e0e611c03fd0a010022373d021864aba56583c796bf9131c804a2ea40acede728b279af38b48dfdd2fd2501005e5f986a28e1cdf2b55b6e5bbcfa34742c45e016f7a920518f376c4b0cbfa868fd7201001287267e0f74f28a0dc5e3e0376fbf28c5ab06424a4dcfd02bb7a65b62d9849d038400a8109eb92b03a106ef15c9d120d7c34ff07ac280636632561e42280499f020659300a60aa07079a19a3600e7fc87cc6a72455f0c2f2735dd3d4039bdaf498469c4d3b800685a26978dfc493d1e03efd9a1c9cac0122d6d0bf027b523be85a4e2a2df3df90343006f78c5d6f4372c65cad0546446bd893db8c47f65e3eae2803ced606f4fd272924800b4a9bae40c785222f38e4127a169fbcbe45085b3e9c59d9631032d5a48dab2a35d00559d25e90b990524eb274251a22508ae04b36125dc894c4e1be21c43c19ab93c03200046bc3f2f79d2aa7da31093e690bb6c10a011a67f2a382937c5eaf423b903df5325006daf88be61cc906f104ac405b04d19f4771a63857a25915e376b53250abe112e2f0037461e9fa1f435caca254303b400b21cc452343572a68ad80d9d3287c2bd8f0f031100f14eebaa20670ebb3d9dab73d074550e3a93cfcb29d90c56dcc205aa8b6a51ab1300597b51a0440a0afe4c346f3b89c5f7aaa7478449f3eb6283e1e1f55b24e54b3b16001009f8ce41536d05ace952e35ce67fc94da2e97b6550f55fbc1d5aa5f3266829030800b668767e12637b80a04de0decb4b96b980b19bd0480557adebfc0c6a46cff1140900d24cec3667bedd9ff7e8bc26dff6ec5fcf8af5cc09f500cad08fdfa2ab2ccf870a00ebea0722a541a4f9e7c4659fdcad062e5806b8abba40cba82eba6882896a763d02040036de4d36e7fc3f273ddd83171a030a19c8668a1f5e03dd62ad53866f3afc127705005c9dc967c9a6dd0dd9c80660dd8e86faa3d7f070ed086f1b2d137147d0f52af00103008e9919d62be144a097dd23e1bf924b2e468022c12ccf50db6ceab3d043cdfd8d010000ae98483a460252d92b031d49f591d571e29f1c8b0ae9e2596e4cd24b1c549d3c0401000000019ed68f94dfa952554d777dbaa9e5c01acb3df767e40cabad7b6fb7547bfa871a010000006a4730440220287534d6ff51166e014ad91a2b677be4bd88cf08785624006cdb66553eafc8cf02204862f38e9d2982a5ee95a7850222f2208bff38637349ecfe41abe185498e4ead4121035ca1a2c6d2b46c61fd29e7697018f5ce2bae1ae735e23627046a2dd17ca8fb24ffffffff02de000000000000001976a914f5c9505bf02a4a2fb591e3568183f9c53cf157be88aca62b0000000000001976a91489b5e639bce3209e0888ea8b7eb4203de1c6148888ac000000000100010000000154aa46f1b3b7bde36c02e293b74d53e6c6eaed7411d286183b1dca766f42879a010000006b483045022100cd21d346073b4a0788018ff6938c44395d14cf5759fcc35a0899a8fe35a3c2a0022064eb9a005c3d0be03b61ab0e1c8757ed566dd935dacac37fcd1452adba4994b541210272d67492c31d0e6bead28c934fb1c9bb50ba9b46f886209fe95fb6a3e43bb27bffffffff0257040000000000001976a9140501308b6409cca5a7b5768c18ff2de8da4c1fa388ac39420000000000001976a91417e3d89f4aeacd5b4929fe04edc32c79b6182e1988ac0000000001000100000001e230ab1b300ac3ce334590fc308fee93ddbb252f6e4645e0a20f7e30dd541289010000006b483045022100a611fdf01eca42289d80e1265584e5bd487faa72e6142ebbc140a676f7c5037c0220409282aaadf580f458d97d61db43c94ac343e0b40674a80fd3ac47f43fd0c66c4121020a87e70cc26f7d5fe775f622d2705f27cfd6f5d2b574fea75401d6412a58b91affffffff02d2040000000000001976a9145d2117c4f66bdb335ce2707a74c46fa46d02cdb388acf23b0000000000001976a914effd80ee9df812990a8d7834fa8610491cbeb91688ac0000000001000100000003e01e74cab9a0571ab5a7d86794826f756a9c65dd0dea3bb3720c4051c488cf50000000006b483045022100bc7fc6ace1a5b1ab8601599d56b3adad4a11b7f11757f3225e96b46ca1ab7f7c0220324d6074aa987a7c63c404ac5b03c26e55d3c4209e298b4ca9df0e90aca43ef3412103ee05b34332b5662830c600b73f9c908bb8bff1813bc9b2690e9cad00fad23d3cffffffff08c461a39a8877db46472f5cc59e5a108e417b1c9ea3091b71b65346d218f471000000006b483045022100a936c496423ec03b1ad0f3bfe2348572d7b29ab14e4435c0c8e2ee093d930fde02203d9e86647ea18043c150289f74c6cf2ceb9ca3b228ae31c7b19c4eef813fb68d412103a19014bcc672ccdf18abb6972dd699367baed89c29b704385253ce2ae0eddad5ffffffff2256c94d07451664749e440f55cec8a37da1c46cf30a97579e2f9696b84ad484000000006b48304502210091b0bcf2e84d9ee65de437e8396b379941345e4cffac331af2ae29b8a16968a602205a00eed18a7ffe36f59ae6eb477d9002324cfc249c875260e6ade5bce852692d4121021446bd1df2b61952088a22a516550e43cd95e47ca2a778822d21268bd8b1cebeffffffff02c4090000000000001976a91497ebeffef6d9dd88ffbce922f1df97cbcd7f88d388ac42000000000000001976a91449457f2c101859d1c8ff90096385d3cc30e5488388ac0000000000",
	3: "0100beef02fe73780c000c02cc005d2fa529262bd8c5451e2e9e91ff0cffa1100b144bddbb537db6cef3868d68cccd02a114c7deb8deba851d87755aa10aa18c97bd77afee4e1bad01d1c50e07a644eb016700672bf5eb61032d3a9404ff6b2045a7ee3444a0612dd5f810fbf8ea363aae0b0b013200f62a9a8d99f570bfa8eb48ef0da91c84ab8110f826f67b72db053513b78baf430118000fec193c7d2ab70f1c74b7b2f4413db12a2da66cf40630ea0e57884d885dbf18010d0027f259788828eb8461da4e67e89c321c14be138576b039663da2dd84f84bcffc010700ebf4dd6d5eb43dcf25ef8025b17072ad3de6db031866c94570d86696bdb9533301020068d3cb01545e515692b45db66d8cf9f5020b0473fda254b2adc45308eee8db3801000072913db72c489f740085cb7aabe3c31271a2ea908b0c8498a75e89cab82a54c0010100491b7787ad57f9b034c950466026e844826a31d1227131dd54ffe965b3612bc20101001f83fda6372ec5bc5f4fb776506ebc01c14841d5173c4444cc3f37ab0539e00d010100cebf33ed7c7afce3c7c873d863ebccfc98a9450cefef843789c3683ab4dbf2c1010100dc365f1d99781e5053f678bb9e06d4a17bc6ed9a2258e2c69a2908800e4a7f2afea6790c000f02fda82c000703145efd05fec7c2edc1827ec685755a2d05208486e9bc2461268e0f5e533bfda92c02cb3553424ffc94b59a60fb358b6cb6dfb694aee894dcd1effc0ed0a9052464e301fd551600b7b53a09331453b5966589dc473b45c87823f109417593a82ff8ffe7ddc6f96e01fd2b0b0080085c6f18f35d5f0a231eda36c1be3f031734bb6cb6987978ef2aad007ea6da01fd940500d26d24b44097ed3f0c2927413dd2f1fb78bba948803abdb7f2fb51d9807a77bc01fdcb0200bad82dfd55455709713ea8390a7c76be5c076da9cd487b75f558728ef8572b7601fd64010022553e159788764f6b3c1e27324f99a13abee9c7069ce34b8a4fcabc45b7aabb01b300debfa516f54f4331ddf74067403d8a70915973e8300c298ff8dbe5bdcc94768101580079a933dbaeaee5dfb6af1ce9bd7a9ef40584e1844a938e398d19f94aba525bed012d005287b1c986e6495d00c103082618d9e0a30c0c620fd2d232eb9c31d59dec510e01170074639f74ebdaaac679e8fed504ae62a6f42c633a549a99e21940aff14e69ab08010a008eb3fde752d9e5c8e67c26daa95f8b9480ec11e608de7afae04eb19775f42342010400eea8e24927e93a0cf9ca50b315e48efca1f0b77643f7110c6efd94220dcedab6010300682c5f78d5f89e5e5d13619f543aacb00d1e4e85043f4d453bd6f6eb14755e790100006df75ca701801279bcb9b91579ac74b2131913ba35a0c3cb313c49062c8f453f0101009fe83bf3febfc2af4641b9bc8c1f76ff9b2ae7de9f4ab5374e672ac70cf8b9b904010000000150965003ea3d2c08bc79b116c9ffe7e730c9f9cf0a61e3df07868b24eac6f8d3000000006b4830450221009d3489f9e76ff3b043708972c52f85519e50a5fc35563d405e04b668780bf2ba0220024188508fc9c6870b2fc4f40b9484ae4163481199a5b4a7a338b86ec8952fee4121036a8b9d796ce2dee820d1f6d7a6ba07037dab4758f16028654fe4bc3a5c430b40ffffffff022a200000000000001976a91484c73348a8fbbc44cfa34f8f5441fc104f3bc78588ac162f0000000000001976a914590b1df63948c2c4e7a12a6e52012b36e25daa9888ac0000000001000100000001a114c7deb8deba851d87755aa10aa18c97bd77afee4e1bad01d1c50e07a644eb010000006a473044022041abd4f93bd1db1d0097f2d467ae183801d7842d23d0605fa9568040d245167402201be66c96bef4d6d051304f6df2aecbdfe23a8a05af0908ef2117ab5388d8903c412103c08545a40c819f6e50892e31e792d221b6df6da96ebdba9b6fe39305cc6cc768ffffffff0263040000000000001976a91454097d9d921f9a1f55084a943571d868552e924f88acb22a0000000000001976a914c36b3fca5159231033f3fbdca1cde942096d379f88ac00000000000100000001cfc39e3adcd58ed58cf590079dc61c3eb6ec739abb7d22b592fb969d427f33ee000000006a4730440220253e674e64028459457d55b444f5f3dc15c658425e3184c628016739e4921fd502207c8fe20eb34e55e4115fbd82c23878b4e54f01f6c6ad0811282dd0b1df863b5e41210310a4366fd997127ad972b14c56ca2e18f39ca631ac9e3e4ad3d9827865d0cc70ffffffff0264000000000000001976a914668a92ff9cb5785eb8fc044771837a0818b028b588acdc4e0000000000001976a914b073264927a61cf84327dea77414df6c28b11e5988ac0000000001010100000002cb3553424ffc94b59a60fb358b6cb6dfb694aee894dcd1effc0ed0a9052464e3000000006a4730440220515c3bf93d38fa7cc164746fae4bec8b66c60a82509eb553751afa5971c3e41d0220321517fd5c997ab5f8ef0e59048ce9157de46f92b10d882bf898e62f3ee7343d4121038f1273fcb299405d8d140b4de9a2111ecb39291b2846660ebecd864d13bee575ffffffff624fbcb4e68d162361f456b8b4fef6b9e7943013088b32b6bca7f5ced41ff004010000006a47304402203fb24f6e00a6487cf88a3b39d8454786db63d649142ea76374c2f55990777e6302207fbb903d038cf43e13ffb496a64f36637ec7323e5ac48bb96bdb4a885100abca4121024b003d3cf49a8f48c1fe79b711b1d08e306c42a0ab8da004d97fccc4ced3343affffffff026f000000000000001976a914f232d38cd4c2f87c117af06542b04a7061b6640188aca62a0000000000001976a9146058e52d00e3b94211939f68cc2d9a3fc1e3db0f88ac0000000000",
	4: "0100beef02fedd770c000b025a0250965003ea3d2c08bc79b116c9ffe7e730c9f9cf0a61e3df07868b24eac6f8d35b00d23a2585c30d0aa45dc2da97ea2da65d6e8a89600d719ada673861ab65ad525b012c000c4ffef5e63cde231502a66acf392337114b6daea97ca2f7ca5cf6a7be38c384011700dd80513a91875591df2265dae8cdc909a01113459ad6d8d9a5c051e092f26058010a0078ef0896fb0077e619d1a4942ba6272b23ff2f48d345826aaf6393688ab03f6301040052ebb99aa1adce1852250a49fd22e18a1798d08840eb13c3b8b1eb0c13bb804b010300cf10a22ebde10e1d98515d5288305e027e3cc0a28cf24c8b3c22a3dd35033dcf0100005bd338bdd23417971c877571c083b8f7fb2a7c0b179b77bed1ab30252c46c7990101007f673d90f535d5875ba5694d23d2a9653b4ad4ca3ef4e61b1b282654e3c7669a0101005b544e50fd2f878b1d7cc0b5c7b63377552e4ad7f01e5dada39984016141e3400101008713a257ca2ba89608cc62133eb6a3eeab15aec4f2e3d7b1a23e4d92d7b1e6dc0101007ccb140e81af58f6a301bd798024ee9515a53094c58b32dab99c18972b4efee1fea6790c000f02fda82c000703145efd05fec7c2edc1827ec685755a2d05208486e9bc2461268e0f5e533bfda92c02cb3553424ffc94b59a60fb358b6cb6dfb694aee894dcd1effc0ed0a9052464e301fd551600b7b53a09331453b5966589dc473b45c87823f109417593a82ff8ffe7ddc6f96e01fd2b0b0080085c6f18f35d5f0a231eda36c1be3f031734bb6cb6987978ef2aad007ea6da01fd940500d26d24b44097ed3f0c2927413dd2f1fb78bba948803abdb7f2fb51d9807a77bc01fdcb0200bad82dfd55455709713ea8390a7c76be5c076da9cd487b75f558728ef8572b7601fd64010022553e159788764f6b3c1e27324f99a13abee9c7069ce34b8a4fcabc45b7aabb01b300debfa516f54f4331ddf74067403d8a70915973e8300c298ff8dbe5bdcc94768101580079a933dbaeaee5dfb6af1ce9bd7a9ef40584e1844a938e398d19f94aba525bed012d005287b1c986e6495d00c103082618d9e0a30c0c620fd2d232eb9c31d59dec510e01170074639f74ebdaaac679e8fed504ae62a6f42c633a549a99e21940aff14e69ab08010a008eb3fde752d9e5c8e67c26daa95f8b9480ec11e608de7afae04eb19775f42342010400eea8e24927e93a0cf9ca50b315e48efca1f0b77643f7110c6efd94220dcedab6010300682c5f78d5f89e5e5d13619f543aacb00d1e4e85043f4d453bd6f6eb14755e790100006df75ca701801279bcb9b91579ac74b2131913ba35a0c3cb313c49062c8f453f0101009fe83bf3febfc2af4641b9bc8c1f76ff9b2ae7de9f4ab5374e672ac70cf8b9b9050100000002787a565270ec00b1bf6ed20100223176656705dc0cfe5ef9d1810ca6569f12d1020000006a47304402203cfe36be7ff5c2ac939bb6a625e4a1226be242f1f9950672b5f696ec58a3358902202a48d6c6e81e5950dc49d0dd1a35b46fa8f919b109b0e7c05deaef3db6051890412102fb130326dbd7c43841cde467196e5f289b9d8596e237725df84f768468426d8bffffffff008d9db2a5c8c310e6394c24c1f3c23b3adbdd6ab4a719e917a4a0ed78768773020000006a473044022049c80385f7f69e8ba6039ebe84fe5e6578f4c3c83eb622442a96219c59ac1a750220317fe2b47838dff11f88d909732d0846eba20acff57cb357a3ff39b5a7b61b3741210322b79b40a759c485eac318eabba60a73a49ec3307ded79ba8c47204405bb2f3fffffffff05414f0000000000001976a91400414bcf2602f309171901d837b4a155adbfb5ce88ac50c30000000000001976a91489ef778cc07c77cce1ad3ff6274615afe15f20c088ac204e0000000000001976a914971b76df1dc6acf01e8e7d2f8bfb3c86e69bc64c88acef250000000000001976a9144b4a836b444d5ed8d245ddb1aa878908e36cd6b588ac9d860100000000001976a9144405da67e318e9cfd9d6ce9dffce27af5f60522888ac000000000100010000000150965003ea3d2c08bc79b116c9ffe7e730c9f9cf0a61e3df07868b24eac6f8d3000000006b4830450221009d3489f9e76ff3b043708972c52f85519e50a5fc35563d405e04b668780bf2ba0220024188508fc9c6870b2fc4f40b9484ae4163481199a5b4a7a338b86ec8952fee4121036a8b9d796ce2dee820d1f6d7a6ba07037dab4758f16028654fe4bc3a5c430b40ffffffff022a200000000000001976a91484c73348a8fbbc44cfa34f8f5441fc104f3bc78588ac162f0000000000001976a914590b1df63948c2c4e7a12a6e52012b36e25daa9888ac00000000000100000001a114c7deb8deba851d87755aa10aa18c97bd77afee4e1bad01d1c50e07a644eb010000006a473044022041abd4f93bd1db1d0097f2d467ae183801d7842d23d0605fa9568040d245167402201be66c96bef4d6d051304f6df2aecbdfe23a8a05af0908ef2117ab5388d8903c412103c08545a40c819f6e50892e31e792d221b6df6da96ebdba9b6fe39305cc6cc768ffffffff0263040000000000001976a91454097d9d921f9a1f55084a943571d868552e924f88acb22a0000000000001976a914c36b3fca5159231033f3fbdca1cde942096d379f88ac00000000000100000001cfc39e3adcd58ed58cf590079dc61c3eb6ec739abb7d22b592fb969d427f33ee000000006a4730440220253e674e64028459457d55b444f5f3dc15c658425e3184c628016739e4921fd502207c8fe20eb34e55e4115fbd82c23878b4e54f01f6c6ad0811282dd0b1df863b5e41210310a4366fd997127ad972b14c56ca2e18f39ca631ac9e3e4ad3d9827865d0cc70ffffffff0264000000000000001976a914668a92ff9cb5785eb8fc044771837a0818b028b588acdc4e0000000000001976a914b073264927a61cf84327dea77414df6c28b11e5988ac0000000001010100000002cb3553424ffc94b59a60fb358b6cb6dfb694aee894dcd1effc0ed0a9052464e3000000006a4730440220515c3bf93d38fa7cc164746fae4bec8b66c60a82509eb553751afa5971c3e41d0220321517fd5c997ab5f8ef0e59048ce9157de46f92b10d882bf898e62f3ee7343d4121038f1273fcb299405d8d140b4de9a2111ecb39291b2846660ebecd864d13bee575ffffffff624fbcb4e68d162361f456b8b4fef6b9e7943013088b32b6bca7f5ced41ff004010000006a47304402203fb24f6e00a6487cf88a3b39d8454786db63d649142ea76374c2f55990777e6302207fbb903d038cf43e13ffb496a64f36637ec7323e5ac48bb96bdb4a885100abca4121024b003d3cf49a8f48c1fe79b711b1d08e306c42a0ab8da004d97fccc4ced3343affffffff026f000000000000001976a914f232d38cd4c2f87c117af06542b04a7061b6640188aca62a0000000000001976a9146058e52d00e3b94211939f68cc2d9a3fc1e3db0f88ac0000000000",
}
