import express from "express";
import * as mainnet from "mainnet-js";

const app = express();

app.use(express.json());

app.post("/verify", async (req, res) => {
    try {
        const {
            address,
            message,
            signature
        } = req.body;

        const verification = await mainnet.SignedMessage.verify(
            message,
            signature,
            address
        );

        res.json({
            valid: verification.valid === true
        });
    } catch (err) {
        res.status(500).json({
            valid: false,
            error: err.message
        });
    }
});


app.post("/token-payment", async (req, res) => {
    try {
        const { address, category, amount } = req.body;
        if (!address || !category || amount === undefined) {
            res.status(400).json({ paid: false, error: "address, category, and amount are required" });
            return;
        }

        const required = BigInt(String(amount));
        const watch = await mainnet.WatchWallet.watchOnly(address);
        await watch.waitForUpdate({ timeout: 10000 }).catch(() => null);
        const utxos = await watch.getUtxos();

        let total = 0n;
        let txid = "";
        for (const utxo of utxos) {
            if (utxo.token?.category === category) {
                const tokenAmount = BigInt(utxo.token.amount || 0n);
                total += tokenAmount;
                if (!txid && tokenAmount > 0n) txid = utxo.txid || utxo.tx_hash || "";
            }
        }

        await watch.stop().catch(() => null);

        res.json({
            paid: total >= required,
            txid,
            token_amount: total.toString(),
            required: required.toString()
        });
    } catch (err) {
        res.status(500).json({ paid: false, error: err.message });
    }
});

app.listen(8788, () => {
    console.log("Verifier running on :8788");
});