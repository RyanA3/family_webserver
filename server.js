require("dotenv").config();
const is_production = process.env.NODE_ENV === "production";
const root_directory = process.env.ROOT_DIR;

const pug = require('pug');

const express = require('express');
const app = express();
const port = 3000;

const pages = require('./routes/pages.js');
const images = require('./routes/images.js');



const mongoose = require('mongoose');
mongoose.connect(process.env.MONGO_AUTH_URL)
console.log('Database connected');



var index = pug.compileFile(`${root_directory}/views/index.pug`);
app.get('/', (req, res) => {
    if(!is_production) index = pug.compileFile(`${root_directory}/views/index.pug`);
    res.send(index());
})

app.use('/page', pages);
app.use('/image', images);



const server = app.listen(port, () => {
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