const fs = require("fs");
const path = require("path");

const pkg = require("../package.json");
const v = pkg.version;
const cargoPath = path.join(__dirname, "../src-tauri/Cargo.toml");
const confPath = path.join(__dirname, "../src-tauri/tauri.conf.json");

let cargo = fs.readFileSync(cargoPath, "utf8");
cargo = cargo.replace(/version\s*=\s*"[^"]+"/m, `version = "${v}"`);
fs.writeFileSync(cargoPath, cargo);

const conf = JSON.parse(fs.readFileSync(confPath, "utf8"));
conf.version = v;
fs.writeFileSync(confPath, JSON.stringify(conf, null, 2) + "\n");

console.log("Synced version to", v);
