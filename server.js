require("dotenv").config();
const is_production = process.env.NODE_ENV === "production";
const root_directory = process.env.ROOT_DIR;

var path = require('path')

const pug = require('pug');

const express = require('express');
const app = express();
const port = 3000;

const pages = require('./controllers/pages.js');
const images = require('./controllers/images.js');

const mongoose = require('mongoose');

mongoose.connect(process.env.MONGO_AUTH_URL, {
    dbName: process.env.MONGO_DATABASE
})
console.log('Database connected');

const ImageMeta = require("./models/ImageMeta.js")



var index = pug.compileFile(`${root_directory}/views/index.pug`);
app.get('/', (req, res) => {
    if(!is_production) index = pug.compileFile(`${root_directory}/views/index.pug`);
    res.send(index());
})

app.use('/page', pages);
app.use('/image', images);

app.use(express.static(process.env.FILES_DIR))



const server = app.listen(port, '0.0.0.0', () => {
    console.log(`Example app listening on port ${port}`)
})



//Handle server shutdown
process.on('SIGTERM', () => {
    console.log('Shutting down...');
    server.close(() => {
        console.log('Closed http server');

        mongoose.connection.close(false, () => {
            console.log('Database connection closed');
            process.exit(0);
        })
    });
})