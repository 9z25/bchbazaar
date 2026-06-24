{
  "contractName": "BCHBazaarEscrow",
  "constructorInputs": [
    { "name": "sellerPkh", "type": "bytes20" },
    { "name": "buyerPkh", "type": "bytes20" },
    { "name": "moderatorPkh", "type": "bytes20" },
    { "name": "feePkh", "type": "bytes20" },
    { "name": "tokenCategory", "type": "bytes" },
    { "name": "refundLocktime", "type": "int" },
    { "name": "amount", "type": "int" },
    { "name": "feeAmount", "type": "int" },
    { "name": "tokenDust", "type": "int" }
  ],
  "abi": [
    {
      "name": "sellerClaim",
      "inputs": [
        { "name": "sellerPubkey", "type": "pubkey" },
        { "name": "sellerSig", "type": "sig" }
      ]
    },
    {
      "name": "buyerRefundAfterTimeout",
      "inputs": [
        { "name": "buyerPubkey", "type": "pubkey" },
        { "name": "buyerSig", "type": "sig" }
      ]
    },
    {
      "name": "moderatorReleaseToSeller",
      "inputs": [
        { "name": "moderatorPubkey", "type": "pubkey" },
        { "name": "moderatorSig", "type": "sig" }
      ]
    },
    {
      "name": "moderatorRefundToBuyer",
      "inputs": [
        { "name": "moderatorPubkey", "type": "pubkey" },
        { "name": "moderatorSig", "type": "sig" }
      ]
    }
  ],
  "bytecode": "OP_9 OP_PICK OP_0 OP_NUMEQUAL OP_IF OP_10 OP_PICK OP_HASH160 OP_OVER OP_EQUALVERIFY OP_11 OP_ROLL OP_11 OP_ROLL OP_CHECKSIGVERIFY OP_4 OP_PICK OP_0 OP_EQUAL OP_IF OP_0 OP_OUTPUTBYTECODE 76a914 OP_2 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_0 OP_OUTPUTVALUE OP_7 OP_PICK OP_9 OP_PICK OP_SUB OP_GREATERTHANOREQUAL OP_VERIFY OP_7 OP_PICK OP_0 OP_GREATERTHAN OP_IF OP_1 OP_OUTPUTBYTECODE 76a914 OP_5 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_1 OP_OUTPUTVALUE OP_8 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_ENDIF OP_ELSE OP_0 OP_OUTPUTBYTECODE 76a914 OP_2 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_0 OP_OUTPUTVALUE OP_9 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_0 OP_OUTPUTTOKENCATEGORY OP_5 OP_PICK OP_EQUALVERIFY OP_0 OP_OUTPUTTOKENAMOUNT OP_7 OP_PICK OP_9 OP_PICK OP_SUB OP_GREATERTHANOREQUAL OP_VERIFY OP_7 OP_PICK OP_0 OP_GREATERTHAN OP_IF OP_1 OP_OUTPUTBYTECODE 76a914 OP_5 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_1 OP_OUTPUTVALUE OP_9 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_1 OP_OUTPUTTOKENCATEGORY OP_5 OP_PICK OP_EQUALVERIFY OP_1 OP_OUTPUTTOKENAMOUNT OP_8 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_ENDIF OP_ENDIF OP_2DROP OP_2DROP OP_2DROP OP_2DROP OP_2DROP OP_1 OP_ELSE OP_9 OP_PICK OP_1 OP_NUMEQUAL OP_IF OP_10 OP_PICK OP_HASH160 OP_2 OP_PICK OP_EQUALVERIFY OP_11 OP_ROLL OP_11 OP_ROLL OP_CHECKSIGVERIFY OP_5 OP_ROLL OP_CHECKLOCKTIMEVERIFY OP_DROP OP_4 OP_PICK OP_0 OP_EQUAL OP_IF OP_0 OP_OUTPUTBYTECODE 76a914 OP_3 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_0 OP_OUTPUTVALUE OP_6 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_ELSE OP_0 OP_OUTPUTBYTECODE 76a914 OP_3 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_0 OP_OUTPUTVALUE OP_8 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_0 OP_OUTPUTTOKENCATEGORY OP_5 OP_PICK OP_EQUALVERIFY OP_0 OP_OUTPUTTOKENAMOUNT OP_6 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_ENDIF OP_2DROP OP_2DROP OP_2DROP OP_2DROP OP_DROP OP_1 OP_ELSE OP_9 OP_PICK OP_2 OP_NUMEQUAL OP_IF OP_10 OP_PICK OP_HASH160 OP_3 OP_ROLL OP_EQUALVERIFY OP_10 OP_ROLL OP_10 OP_ROLL OP_CHECKSIGVERIFY OP_3 OP_PICK OP_0 OP_EQUAL OP_IF OP_0 OP_OUTPUTBYTECODE 76a914 OP_2 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_0 OP_OUTPUTVALUE OP_6 OP_PICK OP_8 OP_PICK OP_SUB OP_GREATERTHANOREQUAL OP_VERIFY OP_6 OP_PICK OP_0 OP_GREATERTHAN OP_IF OP_1 OP_OUTPUTBYTECODE 76a914 OP_4 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_1 OP_OUTPUTVALUE OP_7 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_ENDIF OP_ELSE OP_0 OP_OUTPUTBYTECODE 76a914 OP_2 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_0 OP_OUTPUTVALUE OP_8 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_0 OP_OUTPUTTOKENCATEGORY OP_4 OP_PICK OP_EQUALVERIFY OP_0 OP_OUTPUTTOKENAMOUNT OP_6 OP_PICK OP_8 OP_PICK OP_SUB OP_GREATERTHANOREQUAL OP_VERIFY OP_6 OP_PICK OP_0 OP_GREATERTHAN OP_IF OP_1 OP_OUTPUTBYTECODE 76a914 OP_4 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_1 OP_OUTPUTVALUE OP_8 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_1 OP_OUTPUTTOKENCATEGORY OP_4 OP_PICK OP_EQUALVERIFY OP_1 OP_OUTPUTTOKENAMOUNT OP_7 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_ENDIF OP_ENDIF OP_2DROP OP_2DROP OP_2DROP OP_2DROP OP_DROP OP_1 OP_ELSE OP_9 OP_ROLL OP_3 OP_NUMEQUALVERIFY OP_9 OP_PICK OP_HASH160 OP_3 OP_ROLL OP_EQUALVERIFY OP_9 OP_ROLL OP_9 OP_ROLL OP_CHECKSIGVERIFY OP_3 OP_PICK OP_0 OP_EQUAL OP_IF OP_0 OP_OUTPUTBYTECODE 76a914 OP_3 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_0 OP_OUTPUTVALUE OP_6 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_ELSE OP_0 OP_OUTPUTBYTECODE 76a914 OP_3 OP_PICK OP_CAT 88ac OP_CAT OP_EQUALVERIFY OP_0 OP_OUTPUTVALUE OP_8 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_0 OP_OUTPUTTOKENCATEGORY OP_4 OP_PICK OP_EQUALVERIFY OP_0 OP_OUTPUTTOKENAMOUNT OP_6 OP_PICK OP_GREATERTHANOREQUAL OP_VERIFY OP_ENDIF OP_2DROP OP_2DROP OP_2DROP OP_2DROP OP_1 OP_ENDIF OP_ENDIF OP_ENDIF",
  "source": "pragma cashscript ^0.13.0;\r\n\r\ncontract BCHBazaarEscrow(\r\n    bytes20 sellerPkh,\r\n    bytes20 buyerPkh,\r\n    bytes20 moderatorPkh,\r\n    bytes20 feePkh,\r\n\r\n    // 0x for BCH orders.\r\n    // PUSD token category for PUSD orders.\r\n    bytes tokenCategory,\r\n\r\n    int refundLocktime,\r\n\r\n    // BCH order: amount = sats\r\n    // PUSD order: amount = token base units\r\n    int amount,\r\n\r\n    // BCH order: feeAmount = sats\r\n    // PUSD order: feeAmount = token base units\r\n    int feeAmount,\r\n\r\n    // Dust used for token outputs.\r\n    // For BCH orders this can still be 0.\r\n    int tokenDust\r\n) {\r\n    function sellerClaim(\r\n        pubkey sellerPubkey,\r\n        sig sellerSig\r\n    ) {\r\n        require(hash160(sellerPubkey) == sellerPkh);\r\n        require(checkSig(sellerSig, sellerPubkey));\r\n\r\n        if (tokenCategory == 0x) {\r\n            // BCH order\r\n            require(tx.outputs[0].lockingBytecode == new LockingBytecodeP2PKH(sellerPkh));\r\n            require(tx.outputs[0].value >= amount - feeAmount);\r\n\r\n            if (feeAmount > 0) {\r\n                require(tx.outputs[1].lockingBytecode == new LockingBytecodeP2PKH(feePkh));\r\n                require(tx.outputs[1].value >= feeAmount);\r\n            }\r\n        } else {\r\n            // PUSD / CashToken order\r\n\r\n            // output 0 = seller token payment\r\n            require(tx.outputs[0].lockingBytecode == new LockingBytecodeP2PKH(sellerPkh));\r\n            require(tx.outputs[0].value >= tokenDust);\r\n            require(tx.outputs[0].tokenCategory == tokenCategory);\r\n            require(tx.outputs[0].tokenAmount >= amount - feeAmount);\r\n\r\n            // output 1 = marketplace token fee\r\n            if (feeAmount > 0) {\r\n                require(tx.outputs[1].lockingBytecode == new LockingBytecodeP2PKH(feePkh));\r\n                require(tx.outputs[1].value >= tokenDust);\r\n                require(tx.outputs[1].tokenCategory == tokenCategory);\r\n                require(tx.outputs[1].tokenAmount >= feeAmount);\r\n            }\r\n        }\r\n    }\r\n\r\n    function buyerRefundAfterTimeout(\r\n        pubkey buyerPubkey,\r\n        sig buyerSig\r\n    ) {\r\n        require(hash160(buyerPubkey) == buyerPkh);\r\n        require(checkSig(buyerSig, buyerPubkey));\r\n        require(tx.time >= refundLocktime);\r\n\r\n        if (tokenCategory == 0x) {\r\n            // BCH refund\r\n            require(tx.outputs[0].lockingBytecode == new LockingBytecodeP2PKH(buyerPkh));\r\n            require(tx.outputs[0].value >= amount);\r\n        } else {\r\n            // Token refund\r\n            require(tx.outputs[0].lockingBytecode == new LockingBytecodeP2PKH(buyerPkh));\r\n            require(tx.outputs[0].value >= tokenDust);\r\n            require(tx.outputs[0].tokenCategory == tokenCategory);\r\n            require(tx.outputs[0].tokenAmount >= amount);\r\n        }\r\n    }\r\n\r\n    function moderatorReleaseToSeller(\r\n        pubkey moderatorPubkey,\r\n        sig moderatorSig\r\n    ) {\r\n        require(hash160(moderatorPubkey) == moderatorPkh);\r\n        require(checkSig(moderatorSig, moderatorPubkey));\r\n\r\n        if (tokenCategory == 0x) {\r\n            // BCH order\r\n            require(tx.outputs[0].lockingBytecode == new LockingBytecodeP2PKH(sellerPkh));\r\n            require(tx.outputs[0].value >= amount - feeAmount);\r\n\r\n            if (feeAmount > 0) {\r\n                require(tx.outputs[1].lockingBytecode == new LockingBytecodeP2PKH(feePkh));\r\n                require(tx.outputs[1].value >= feeAmount);\r\n            }\r\n        } else {\r\n            // PUSD / CashToken order\r\n\r\n            // output 0 = seller token payment\r\n            require(tx.outputs[0].lockingBytecode == new LockingBytecodeP2PKH(sellerPkh));\r\n            require(tx.outputs[0].value >= tokenDust);\r\n            require(tx.outputs[0].tokenCategory == tokenCategory);\r\n            require(tx.outputs[0].tokenAmount >= amount - feeAmount);\r\n\r\n            // output 1 = marketplace token fee\r\n            if (feeAmount > 0) {\r\n                require(tx.outputs[1].lockingBytecode == new LockingBytecodeP2PKH(feePkh));\r\n                require(tx.outputs[1].value >= tokenDust);\r\n                require(tx.outputs[1].tokenCategory == tokenCategory);\r\n                require(tx.outputs[1].tokenAmount >= feeAmount);\r\n            }\r\n        }\r\n    }\r\n\r\n    function moderatorRefundToBuyer(\r\n        pubkey moderatorPubkey,\r\n        sig moderatorSig\r\n    ) {\r\n        require(hash160(moderatorPubkey) == moderatorPkh);\r\n        require(checkSig(moderatorSig, moderatorPubkey));\r\n\r\n        if (tokenCategory == 0x) {\r\n            // BCH refund\r\n            require(tx.outputs[0].lockingBytecode == new LockingBytecodeP2PKH(buyerPkh));\r\n            require(tx.outputs[0].value >= amount);\r\n        } else {\r\n            // Token refund\r\n            require(tx.outputs[0].lockingBytecode == new LockingBytecodeP2PKH(buyerPkh));\r\n            require(tx.outputs[0].value >= tokenDust);\r\n            require(tx.outputs[0].tokenCategory == tokenCategory);\r\n            require(tx.outputs[0].tokenAmount >= amount);\r\n        }\r\n    }\r\n}",
  "fingerprint": "fcd9509780a6ebc8db15bb0f085003d26df07453ab0f788d698b8ba52feb6206",
  "debug": {
    "bytecode": "5979009c635a79a978885b7a5b7aad547900876300cd0376a91452797e0288ac7e8800cc5779597994a269577900a06351cd0376a91455797e0288ac7e8851cc5879a269686700cd0376a91452797e0288ac7e8800cc5979a26900d155798800d35779597994a269577900a06351cd0376a91455797e0288ac7e8851cc5979a26951d155798851d35879a26968686d6d6d6d6d51675979519c635a79a95279885b7a5b7aad557ab175547900876300cd0376a91453797e0288ac7e8800cc5679a2696700cd0376a91453797e0288ac7e8800cc5879a26900d155798800d35679a269686d6d6d6d7551675979529c635a79a9537a885a7a5a7aad537900876300cd0376a91452797e0288ac7e8800cc5679587994a269567900a06351cd0376a91454797e0288ac7e8851cc5779a269686700cd0376a91452797e0288ac7e8800cc5879a26900d154798800d35679587994a269567900a06351cd0376a91454797e0288ac7e8851cc5879a26951d154798851d35779a26968686d6d6d6d755167597a539d5979a9537a88597a597aad537900876300cd0376a91453797e0288ac7e8800cc5679a2696700cd0376a91453797e0288ac7e8800cc5879a26900d154798800d35679a269686d6d6d6d51686868",
    "sourceMap": "27:4:60:5;;;;;31:24:31:36;;:16::37:1;:41::50:0;:8::52:1;32:25:32:34:0;;:36::48;;:8::51:1;34:12:34:25:0;;:29::31;:12:::1;:33:43:9:0;36:31:36:32;:20::49:1;:53::88:0;:78::87;;:53::88:1;;;:12::90;37:31:37:32:0;:20::39:1;:43::49:0;;:52::61;;:43:::1;:20;:12::63;39:16:39:25:0;;:28::29;:16:::1;:31:42:13:0;40:35:40:36;:24::53:1;:57::89:0;:82::88;;:57::89:1;;;:16::91;41:35:41:36:0;:24::43:1;:47::56:0;;:24:::1;:16::58;39:31:42:13;43:15:59:9:0;47:31:47:32;:20::49:1;:53::88:0;:78::87;;:53::88:1;;;:12::90;48:31:48:32:0;:20::39:1;:43::52:0;;:20:::1;:12::54;49:31:49:32:0;:20::47:1;:51::64:0;;:12::66:1;50:31:50:32:0;:20::45:1;:49::55:0;;:58::67;;:49:::1;:20;:12::69;53:16:53:25:0;;:28::29;:16:::1;:31:58:13:0;54:35:54:36;:24::53:1;:57::89:0;:82::88;;:57::89:1;;;:16::91;55:35:55:36:0;:24::43:1;:47::56:0;;:24:::1;:16::58;56:35:56:36:0;:24::51:1;:55::68:0;;:16::70:1;57:35:57:36:0;:24::49:1;:53::62:0;;:24:::1;:16::64;53:31:58:13;43:15:59:9;30:6:60:5;;;;;;27:4;62::81::0;;;;;66:24:66:35;;:16::36:1;:40::48:0;;:8::50:1;67:25:67:33:0;;:35::46;;:8::49:1;68:27:68:41:0;;:8::43:1;;70:12:70:25:0;;:29::31;:12:::1;:33:74:9:0;72:31:72:32;:20::49:1;:53::87:0;:78::86;;:53::87:1;;;:12::89;73:31:73:32:0;:20::39:1;:43::49:0;;:20:::1;:12::51;74:15:80:9:0;76:31:76:32;:20::49:1;:53::87:0;:78::86;;:53::87:1;;;:12::89;77:31:77:32:0;:20::39:1;:43::52:0;;:20:::1;:12::54;78:31:78:32:0;:20::47:1;:51::64:0;;:12::66:1;79:31:79:32:0;:20::45:1;:49::55:0;;:20:::1;:12::57;74:15:80:9;65:6:81:5;;;;;;62:4;83::116::0;;;;;87:24:87:39;;:16::40:1;:44::56:0;;:8::58:1;88:25:88:37:0;;:39::54;;:8::57:1;90:12:90:25:0;;:29::31;:12:::1;:33:99:9:0;92:31:92:32;:20::49:1;:53::88:0;:78::87;;:53::88:1;;;:12::90;93:31:93:32:0;:20::39:1;:43::49:0;;:52::61;;:43:::1;:20;:12::63;95:16:95:25:0;;:28::29;:16:::1;:31:98:13:0;96:35:96:36;:24::53:1;:57::89:0;:82::88;;:57::89:1;;;:16::91;97:35:97:36:0;:24::43:1;:47::56:0;;:24:::1;:16::58;95:31:98:13;99:15:115:9:0;103:31:103:32;:20::49:1;:53::88:0;:78::87;;:53::88:1;;;:12::90;104:31:104:32:0;:20::39:1;:43::52:0;;:20:::1;:12::54;105:31:105:32:0;:20::47:1;:51::64:0;;:12::66:1;106:31:106:32:0;:20::45:1;:49::55:0;;:58::67;;:49:::1;:20;:12::69;109:16:109:25:0;;:28::29;:16:::1;:31:114:13:0;110:35:110:36;:24::53:1;:57::89:0;:82::88;;:57::89:1;;;:16::91;111:35:111:36:0;:24::43:1;:47::56:0;;:24:::1;:16::58;112:35:112:36:0;:24::51:1;:55::68:0;;:16::70:1;113:35:113:36:0;:24::49:1;:53::62:0;;:24:::1;:16::64;109:31:114:13;99:15:115:9;86:6:116:5;;;;;;83:4;118::136::0;;;;122:24:122:39;;:16::40:1;:44::56:0;;:8::58:1;123:25:123:37:0;;:39::54;;:8::57:1;125:12:125:25:0;;:29::31;:12:::1;:33:129:9:0;127:31:127:32;:20::49:1;:53::87:0;:78::86;;:53::87:1;;;:12::89;128:31:128:32:0;:20::39:1;:43::49:0;;:20:::1;:12::51;129:15:135:9:0;131:31:131:32;:20::49:1;:53::87:0;:78::86;;:53::87:1;;;:12::89;132:31:132:32:0;:20::39:1;:43::52:0;;:20:::1;:12::54;133:31:133:32:0;:20::47:1;:51::64:0;;:12::66:1;134:31:134:32:0;:20::45:1;:49::55:0;;:20:::1;:12::57;129:15:135:9;121:6:136:5;;;;;3:0:137:1;;",
    "logs": [],
    "requires": [
      { "ip": 18, "line": 31 },
      { "ip": 23, "line": 32 },
      { "ip": 37, "line": 36 },
      { "ip": 46, "line": 37 },
      { "ip": 60, "line": 40 },
      { "ip": 66, "line": 41 },
      { "ip": 77, "line": 47 },
      { "ip": 83, "line": 48 },
      { "ip": 88, "line": 49 },
      { "ip": 97, "line": 50 },
      { "ip": 111, "line": 54 },
      { "ip": 117, "line": 55 },
      { "ip": 122, "line": 56 },
      { "ip": 128, "line": 57 },
      { "ip": 148, "line": 66 },
      { "ip": 153, "line": 67 },
      { "ip": 156, "line": 68 },
      { "ip": 171, "line": 72 },
      { "ip": 177, "line": 73 },
      { "ip": 187, "line": 76 },
      { "ip": 193, "line": 77 },
      { "ip": 198, "line": 78 },
      { "ip": 204, "line": 79 },
      { "ip": 223, "line": 87 },
      { "ip": 228, "line": 88 },
      { "ip": 242, "line": 92 },
      { "ip": 251, "line": 93 },
      { "ip": 265, "line": 96 },
      { "ip": 271, "line": 97 },
      { "ip": 282, "line": 103 },
      { "ip": 288, "line": 104 },
      { "ip": 293, "line": 105 },
      { "ip": 302, "line": 106 },
      { "ip": 316, "line": 110 },
      { "ip": 322, "line": 111 },
      { "ip": 327, "line": 112 },
      { "ip": 333, "line": 113 },
      { "ip": 352, "line": 122 },
      { "ip": 357, "line": 123 },
      { "ip": 371, "line": 127 },
      { "ip": 377, "line": 128 },
      { "ip": 387, "line": 131 },
      { "ip": 393, "line": 132 },
      { "ip": 398, "line": 133 },
      { "ip": 404, "line": 134 }
    ]
  },
  "compiler": {
    "name": "cashc",
    "version": "0.13.0",
    "options": {
      "enforceFunctionParameterTypes": true,
      "enforceLocktimeGuard": true
    }
  },
  "updatedAt": "2026-06-22T22:45:28.092Z"
}
