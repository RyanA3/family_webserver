require("dotenv").config();
const pug = require('pug');
const express = require('express');
const app = express();
const port = 3000;
const root_directory = process.env.ROOT_DIR;




app.get('/', (req, res) => {
    const index = pug.compileFile(`${root_directory}/views/index.pug`);
    res.send(index());
})

app.get('/testurlmixin', (req,res) => {
    res.send("<div>Got url data!</div>");
})

app.listen(port, () => {
    console.log(`Example app listening on port ${port}`)
})