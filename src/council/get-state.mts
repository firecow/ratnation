import {State} from "../state-handler.mjs";
import {RouteRes} from "./council-server.mjs";

interface GetStateOpts {state: State; res: RouteRes}

export default async function ({res, state}: GetStateOpts) {
    res.setHeader("Content-Type", "application/json; charset=utf-8");
    res.end(JSON.stringify(state));
    return Promise.resolve();
}
