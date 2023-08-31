import {State} from "../state-handler.js";
import {RouteRes} from "./council-server.js";

interface GetStateOpts {state: State; res: RouteRes}

export default async function ({res, state}: GetStateOpts) {
    res.setHeader("Content-Type", "application/json; charset=utf-8");
    res.end(JSON.stringify(state, null, 2));
    return Promise.resolve();
}
