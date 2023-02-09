import {IncomingMessage, ServerResponse} from "http";
import {State} from "../state-handler.mjs";

export default async function getState (req: IncomingMessage, res: ServerResponse, state: State) {
    res.setHeader("Content-Type", "application/json; charset=utf-8");
    res.end(JSON.stringify(state));
    return Promise.resolve();
}
