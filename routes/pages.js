require("dotenv").config();
const is_production = process.env.NODE_ENV === "production";
const root_directory = process.env.ROOT_DIR;

const pug = require('pug');
const fs = require('fs');

const express = require('express');
const router = express.Router();


//TODO: Make 404 and error pages compile once, call function with proper args to render without compiling every time
const dir_404_page = `${root_directory}/views/pages/404.pug`;
const dir_err_page = `${root_directory}/views/pages/err.pug`;
var page404 = pug.compileFile(dir_404_page);
var pageErr = pug.compileFile(dir_err_page);

router.get('/:page', (req, res) => {
    const target_file = `${root_directory}/views/pages/${req.params.page}.pug`;
    const target_file_exists = fs.existsSync(target_file);

    if(!target_file_exists) {
        if(!is_production) page404 = pug.compileFile(dir_404_page);
        res.send(page404({pageurl: req.params.page}))
        return;
    }

    try {
        const page = pug.compileFile(target_file);
        res.send(page());
        
    } catch (e) {
        if(!is_production) pageErr = pug.compileFile(dir_err_page);
        res.send(pageErr({pageurl: req.params.page, e: e.stack}))
    }
    
});

module.exports = router;