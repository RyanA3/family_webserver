require("dotenv").config();
const pug = require('pug');
const express = require('express');
const app = express();
const port = 3000;
const root_directory = process.env.ROOT_DIR;

function generateLinkCard(url, title) {
    return (
        `<div class="">
            <p>This is a test ${title}</p>
            <a>${url}</a>
        </div>`
    )
}

const compiledHome = pug.compileFile(`${root_directory}/views/index.pug`);

app.get('/', (req, res) => {
    //res.sendFile('/views/index.html', {root: root_directory});
    res.send(compiledHome());
})

app.get('/linkcard/:name/:url', (req, res) => {
    console.log("Request to get linkcard: " + req.params.url + " " + req.params.name)
    //res.send(generateLinkCard(req.params.url, req.params.name));

    var card = pug.compileFile(`${root_directory}/views/components/linkcard.pug`);
    res.send(card({name: req.params.name, url: req.params.url}));
});

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