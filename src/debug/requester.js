import got from "got";
import wait from "wait-promise";
import {to} from "await-to-js";

export const command = "requester";
export const description = "Start calling http requests and print status code";

export async function handler(argv) {
    let shuttingDown = false;
    process.on("SIGINT", () => shuttingDown = true);
    process.on("SIGTERM", () => shuttingDown = true);

    while (!shuttingDown) {
        const [err, res] = await to(got.get(argv["url"], {timeout: {lookup: 100, connect: 100, socket: 100}}));
        if (err) {
            console.error(`message="request failed" ${err.message} service.type=requester`);
            continue;
        }
        if (res.statusCode !== 200) {
            console.error(`message="request failed" http.response.status_code=${res.statusCode} service.type=requester`);
            continue;
        }
        console.log(`message="request succeeded" http.response.status_code=${res.statusCode} service.type=requester`);
        await wait.sleep(argv["sleep"]);
    }
}

export function builder(yargs) {
    yargs.options("url", {
        type: "string",
        description: "URL to request",
        default: "http://localhost:2183",
    });
    yargs.options("sleep", {
        type: "number",
        description: "Time to sleep after each  request",
        default: 500,
    });
}
