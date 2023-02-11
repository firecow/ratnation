import {to} from "await-to-js";
import got from "got";
import {ArgumentsCamelCase, Argv} from "yargs";
import {Logger} from "../logger.mjs";
import {Ticker} from "../ticker.mjs";

export interface RequestArguments {
    "url": string;
    "interval": number;
}

async function tick (logger: Logger, args: RequestArguments) {
    const [err, res] = await to(got.get(args["url"], {timeout: {lookup: 100, connect: 100, socket: 100}}));
    if (err) {
        return logger.error("Request error", {
            "error.message": err?.message,
            "error.stack_trace": err?.stack,
            "service.type": "requester",
        });
    }
    if (res.statusCode !== 200) {
        return logger.error("Response not 200", {
            "http.response.status_code": res?.statusCode,
            "service.type": "requester",
        });
    }
    logger.info("Request success", {
        "http.response.status_code": res?.statusCode,
        "service.type": "requester",
    });
}

export const command = "requester";
export const description = "Start calling http requests and print status code";

export function handler (args: ArgumentsCamelCase) {
    const logger = new Logger();
    const ticker = new Ticker({
        interval: Number(args["interval"]),
        tick: async () => await tick(logger, args as ArgumentsCamelCase<RequestArguments>)
    });
    process.on("SIGINT", () => ticker.stop());
    process.on("SIGTERM", () => ticker.stop());
    ticker.start();
}

export function builder (yargs: Argv) {
    yargs.options("url", {
        type: "string",
        description: "URL to request",
        default: "http://localhost:2183",
    });
    yargs.options("interval", {
        type: "number",
        description: "Ticker interval",
        default: 500,
    });
    return yargs;
}
