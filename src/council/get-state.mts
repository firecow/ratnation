import {IncomingMessage, ServerResponse} from "http";
import {State} from "../state-handler.mjs";
import {Logger} from "../logger.mjs";

export default async function getState (logger: Logger, req: IncomingMessage, res: ServerResponse, state: State) {
    res.setHeader("Content-Type", "application/json; charset=utf-8");
    res.end(JSON.stringify(state));
    return Promise.resolve();
}
