import http from "http";
import findmyway from "find-my-way";
import {Provisioner} from "./provisioner.js";
import putUnderling from "./put-underling.js";
import getState from "./get-state.js";
import putKing from "./put-king.js";
import putKingActive from "./put-king-active.js";

export const command = "council";
export const description = "Start council";

export async function handler(argv) {
    const state = {
        revision: 0,
        services: [],
        kings: [],
    };

    const provisioner = new Provisioner({state});

    const router = findmyway({
        defaultRoute: (req, res) => {
            res.statusCode = 404;
            res.end();
        }
    });
    router.on("GET", "/state", (req, res) => getState(req, res, state));
    router.on("PUT", "/underling", (req, res) => putUnderling(req, res, state));
    router.on("PUT", "/king", (req, res) => putKing(req, res, state));
    router.on("PUT", "/king-active", (req, res) => putKingActive(req, res, state));

    const server = http.createServer((req, res) => router.lookup(req, res));
    server.listen(argv["port"]);
    await new Promise(resolve => server.once("listening", resolve));
    console.log("ratcouncil ready");

    provisioner.start();
}

export function builder(yargs) {
    yargs.options("port", {
        type: "number",
        description: "Webserver listening port",
        default: "8080",
    });
}
