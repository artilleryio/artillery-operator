const express = require("express");
const app = express();
const port = 3000;

app.use(express.json());

app.get("/common", (_, res) => {
    console.log("/common ... hit");
    res.send({route: "/common"});
});

app.get("/average", (_, res) => {
    console.log("/average ... hit");
    res.send({route: "/average"});
});

app.get("/rare", (_, res) => {
    console.log("/rare ... hit");
    res.send({route: "/rare"});
});

app.listen(port, () => {
    console.log(`App listening at http://localhost:${port}`);
});
