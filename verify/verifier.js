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

        const valid = await mainnet.SignedMessage.verify(
            message,
            signature,
            address
        );

        res.json({
            valid
        });
    } catch (err) {
        res.status(500).json({
            valid: false,
            error: err.message
        });
    }
});

app.listen(8788, () => {
    console.log("Verifier running on :8788");
});