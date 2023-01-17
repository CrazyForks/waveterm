const webpack = require('webpack');
const merge = require('webpack-merge');
const common = require('./webpack.common.js');
const path = require('path');
const CopyPlugin = require("copy-webpack-plugin");
const VERSION = "v0.1.1";

var merged = merge.merge(common, {
    mode: "development",
    output: {
        path: path.resolve(__dirname, "dist-dev"),
        filename: "[name].js",
    },
    devtool: "source-map",
    devServer: {
        static: {
            directory: path.join(__dirname, "static"),
        },
        port: 9000,
        headers: {
            'Cache-Control': 'no-store',
        },
    },
    watchOptions: {
        aggregateTimeout: 200,
    },
});

var definePlugin = new webpack.DefinePlugin({
    __SHDEV__: "true",
    __SHVERSION__: JSON.stringify(VERSION),
    __SHBUILD__: JSON.stringify("devbuild"),
});
merged.plugins.push(definePlugin);

module.exports = merged;

