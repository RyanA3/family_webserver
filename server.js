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

app.get('/linkcard/:name/:url', (req, res) => {
    console.log("Request to get linkcard: " + req.params.url + " " + req.params.name)

    var card = pug.compileFile(`${root_directory}/views/components/linkcard.pug`);
    res.send(card({name: req.params.name, url: req.params.url}));
});

app.get('/testurlmixin', (req,res) => {
    res.send("<div>Got url data!</div>");
})

app.get('/index.css', (req, res) => {
    console.log("Sent css file")
    res.sendFile('/assets/css/index.css', {root: root_directory});
})

app.get('/htmx.min.js', (req, res) => {
    console.log("Sent htmx to client");
    res.sendFile('htmx.min.js', {root: root_directory})
})

app.listen(port, () => {
    console.log(`Example app listening on port ${port}`)
})