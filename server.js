require("dotenv").config();
const pug = require('pug');
const fs = require('fs');
const express = require('express');
const app = express();
const port = 3000;
const root_directory = process.env.ROOT_DIR;




app.get('/', (req, res) => {
    const index = pug.compileFile(`${root_directory}/views/index.pug`);
    res.send(index());
})

//TODO: Make 404 and error pages compile once, call function with proper args to render without compiling every time
const dir_404_page = `${root_directory}/views/pages/404.pug`;
const dir_err_page = `${root_directory}/views/pages/err.pug`;
var page404 = pug.compileFile(dir_404_page);
var pageErr = pug.compileFile(dir_err_page);

app.get('/page/:page', (req, res) => {
    const target_file = `${root_directory}/views/pages/${req.params.page}.pug`;
    const target_file_exists = fs.existsSync(target_file);

    if(!target_file_exists) {
        if(process.env.PUG_ENV === "dev") page404 = pug.compileFile(dir_404_page);
        res.send(page404({pageurl: req.params.page}))
        return;
    }

    try {
        const page = pug.compileFile(target_file);
        fs.huh();
        res.send(page());
        
    } catch (e) {
        if(process.env.PUG_ENV === "dev") pageErr = pug.compileFile(dir_err_page);
        res.send(pageErr({pageurl: req.params.page, e: e}))
    }
    
});

app.get('/home', (req,res) => {
    const home = pug.compileFile(`${root_directory}/views/pages/home.pug`);
    res.send(home());
})

app.get('/images', (req,res) => {
    res.send("<div>Image manager</div>");
})

app.listen(port, () => {
    console.log(`Example app listening on port ${port}`)
})