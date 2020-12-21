const path = require("path");
var webpack = require("webpack");
var fs = require("fs");

var nodeModules = {};
fs.readdirSync("./node_modules")
	.filter(function (x) {
		return [".bin"].indexOf(x) === -1;
	})
	.forEach(function (mod) {
		nodeModules[mod] = "commonjs " + mod;
	});

module.exports = {
	target: "node",
	entry: "./index.html",
	output: {
		filename: "index.html",
		path: path.resolve(__dirname, "dist"),
		libraryTarget: "commonjs2"
	},
	externals: nodeModules
};
